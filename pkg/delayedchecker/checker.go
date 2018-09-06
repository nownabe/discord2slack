package delayedchecker

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/mjibson/goon"
	"github.com/nownabe/discord2slack/pkg/entity"
	"github.com/nownabe/discord2slack/pkg/slack"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

const (
	checkGuildKey   = "checkGuild"
	checkChannelKey = "checkChannel"
	messagesLimit   = 100
)

var (
	DiscordToken    string
	SlackWebhookURL string

	delayedCheckGuild   = delay.Func(checkGuildKey, checkGuild)
	delayedCheckChannel = delay.Func(checkChannelKey, checkChannel)
)

func Call(ctx context.Context, guildID string) error {
	return delayedCheckGuild.Call(ctx, guildID)
}

func checkGuild(ctx context.Context, guildID string) error {
	discord, err := discordgo.New("Bot " + DiscordToken)
	if err != nil {
		log.Errorf(ctx, "failed to create discord client: %v", err)
		return err
	}
	discord.Client = urlfetch.Client(ctx)

	guild, err := discord.Guild(guildID)
	if err != nil {
		log.Errorf(ctx, "failed to get discord guild %s: %v", guildID, err)
		return err
	}

	channels, err := discord.GuildChannels(guild.ID)
	if err != nil {
		log.Errorf(ctx, "failed to get discord channels in %s: %v", guild.ID, err)
		return err
	}

	for _, ch := range channels {
		if ch.Type == discordgo.ChannelTypeGuildText {
			if err := delayedCheckChannel.Call(ctx, guild, ch); err != nil {
				log.Errorf(ctx, "failed to call checkChannel %s in %s: %v", ch.ID, guild.ID, err)
			}
		}
	}

	return nil
}

func checkChannel(ctx context.Context, g *discordgo.Guild, c *discordgo.Channel) error {
	discord, err := discordgo.New("Bot " + DiscordToken)
	if err != nil {
		log.Errorf(ctx, "failed to create discord client: %v", err)
		return err
	}
	discord.Client = urlfetch.Client(ctx)

	slackClient, err := slack.New(SlackWebhookURL)
	if err != nil {
		log.Errorf(ctx,
			"[%s - %s] failed to get slack client: %v",
			g.Name, c.Name, err)
		return err
	}

	guild := &entity.Guild{ID: g.ID, Name: g.Name}
	channel := &entity.Channel{ID: c.ID, Name: c.Name}

	if err := getChannel(ctx, guild, channel); err != nil {
		log.Errorf(ctx,
			"[%s - %s] failed to get channel: %v",
			guild.Name, channel.Name, err)
		return err
	}

	messages, err := discord.ChannelMessages(
		channel.ID, messagesLimit, "", channel.LastMessageID, "")
	if err != nil {
		log.Errorf(ctx,
			"[%s - %s] faild to get discord messages: %v",
			guild.Name, channel.Name, err)
		return err
	}

	if len(messages) == 0 {
		log.Infof(ctx, "[%s - %s] no new messages", guild.Name, channel.Name)
		return nil
	}

	attachments := make([]slack.Attachment, len(messages))

	for i, m := range messages {
		attachments[len(attachments)-i-1] = slack.Attachment{
			AuthorName: m.Author.Username,
			Text:       m.Content,
		}
	}

	text := fmt.Sprintf("New messages in %s - #%s", guild.Name, channel.Name)
	msg := slack.Message{Text: text, Attachments: attachments}
	if err := slackClient.Post(ctx, msg); err != nil {
		log.Errorf(ctx,
			"[%s - %s] failed to post to slack: %v",
			guild.Name, channel.Name, err)
		return err
	}

	channel.LastMessageID = messages[0].ID

	if err := saveChannel(ctx, channel); err != nil {
		log.Errorf(ctx,
			"[%s - %s] failed to save updates: %v",
			guild.Name, channel.Name, err)
		return err
	}

	return nil
}

func getChannel(ctx context.Context, gu *entity.Guild, c *entity.Channel) error {
	g := goon.FromContext(ctx)

	if err := g.Get(gu); err != nil {
		log.Infof(ctx, "guild %s does not exist yet: %v", gu.Name, err)
		if _, err := g.Put(gu); err != nil {
			log.Errorf(ctx, "faild to put build %s: %v", gu.Name, err)
			return err
		}
		log.Infof(ctx, "created guild %s", gu.Name)
	}

	c.GuildKey = g.Key(gu)

	if err := g.Get(c); err != nil {
		log.Infof(ctx, "channel %s does not exist yet: %v", c.Name, err)
		return nil
	}

	return nil
}

func saveChannel(ctx context.Context, c *entity.Channel) error {
	g := goon.FromContext(ctx)

	if _, err := g.Put(c); err != nil {
		log.Errorf(ctx, "failed to put channel %s: %v", c.Name, err)
		return err
	}

	return nil
}
