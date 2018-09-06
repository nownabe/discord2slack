package delayedchecker

import (
	"bytes"
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/nownabe/discord2slack/pkg/slack"
	"google.golang.org/appengine/delay"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

const (
	checkGuildKey   = "checkGuild"
	checkChannelKey = "checkChannel"
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

func checkChannel(ctx context.Context, guild *discordgo.Guild, channel *discordgo.Channel) error {
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
			guild.Name, channel.Name, err)
		return err
	}

	messages, err := discord.ChannelMessages(channel.ID, 100, "", "", "")
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

	text := new(bytes.Buffer)

	for _, m := range messages {
		text.WriteString(fmt.Sprintf(
			"[%s - #%s] %s: %s\n",
			guild.Name,
			channel.Name,
			m.Author.Username,
			m.Content,
		))
		log.Infof(ctx, "%v", m)
	}

	msg := slack.Message{Text: text.String()}
	if err := slackClient.Post(ctx, msg); err != nil {
		log.Errorf(ctx,
			"[%s - %s]failed to post to slack: %v",
			guild.Name, channel.Name, err)
		return err
	}

	return nil
}
