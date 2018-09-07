package channel

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mjibson/goon"
	"github.com/nownabe/discord2slack/pkg/entity"
	"github.com/nownabe/discord2slack/pkg/slack"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

const (
	messagesLimit = 100
)

type checker struct {
	discord *discordgo.Session
	slack   *slack.Client
	guild   *entity.Guild
	channel *entity.Channel
}

func (h *Handler) newChecker(ctx context.Context, g *entity.Guild, c *entity.Channel) (*checker, error) {
	chk := &checker{
		slack:   h.slack,
		guild:   g,
		channel: c,
	}

	discord, err := discordgo.New("Bot " + h.discordToken)
	if err != nil {
		return chk, chk.wrapError(err, "failed to get discord client")
	}
	discord.Client = urlfetch.Client(ctx)

	chk.discord = discord

	return chk, nil
}

func (chk *checker) wrapError(err error, msg string) error {
	return fmt.Errorf("[%s - %s] %s: %v", chk.guild.Name, chk.channel.Name, msg, err)
}

func (chk *checker) check(ctx context.Context) error {
	if err := chk.fetch(ctx); err != nil {
		return chk.wrapError(err, "failed to get fetch channel")
	}

	messages, err := chk.discord.ChannelMessages(
		chk.channel.ID, messagesLimit, "", chk.channel.LastMessageID, "")
	if err != nil {
		return chk.wrapError(err, "failed to get discord messages")
	}

	if len(messages) == 0 {
		log.Infof(ctx, "[%s - %s] no new messages", chk.guild.Name, chk.channel.Name)
		return nil
	}

	attachments := make([]slack.Attachment, len(messages))

	for i, m := range messages {
		attachments[len(attachments)-i-1] = slack.Attachment{
			AuthorName: m.Author.Username,
			Text:       m.Content,
		}
	}

	text := fmt.Sprintf("New messages in %s - #%s", chk.guild.Name, chk.channel.Name)
	msg := slack.Message{Text: text, Attachments: attachments}
	if err := chk.slack.Post(ctx, msg); err != nil {
		return chk.wrapError(err, "failed to post to slack")
	}

	chk.channel.LastMessageID = messages[0].ID

	if err := chk.update(ctx); err != nil {
		return chk.wrapError(err, "failed to save updates")
	}

	return nil
}

func (chk *checker) fetch(ctx context.Context) error {
	g := goon.FromContext(ctx)

	if err := g.Get(chk.guild); err != nil {
		log.Infof(ctx, "guild %s does not exist yet: %v", chk.guild.Name, err)
		if _, err := g.Put(chk.guild); err != nil {
			log.Errorf(ctx, "faild to put build %s: %v", chk.guild.Name, err)
			return err
		}
		log.Infof(ctx, "created guild %s", chk.guild.Name)
	}

	chk.channel.GuildKey = g.Key(chk.guild.Name)

	if err := g.Get(chk.channel); err != nil {
		log.Infof(ctx, "channel %s does not exist yet: %v", chk.channel.Name, err)
		return nil
	}

	return nil
}

func (chk *checker) update(ctx context.Context) error {
	g := goon.FromContext(ctx)

	if _, err := g.Put(chk.channel); err != nil {
		return err
	}

	return nil
}
