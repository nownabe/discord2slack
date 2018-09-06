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

const funcKey = "check"

var (
	DiscordToken    string
	SlackWebhookURL string

	delayFunc = delay.Func(funcKey, check)
)

func Call(ctx context.Context, channelID string) error {
	return delayFunc.Call(ctx, channelID)
}

func check(ctx context.Context, channelID string) error {
	client := urlfetch.Client(ctx)

	discord, err := discordgo.New("Bot " + DiscordToken)
	if err != nil {
		log.Errorf(ctx, "failed to create discord client: %v", err)
		return err
	}

	discord.Client = client

	slackClient, err := slack.New(SlackWebhookURL)
	if err != nil {
		log.Errorf(ctx, "failed to get slack client: %v", err)
		return err
	}

	text := new(bytes.Buffer)

	ch, err := discord.Channel(channelID)
	if err != nil {
		log.Errorf(ctx, "faild to get discord channel: %v", err)
		return err
	}

	g, err := discord.Guild(ch.GuildID)
	if err != nil {
		log.Errorf(ctx, "failed to get discord guild: %v", err)
		return err
	}

	messages, err := discord.ChannelMessages(channelID, 100, "", "", "")
	if err != nil {
		log.Errorf(ctx, "faild to get discord messages at %s in %s: %v", ch.Name, g.Name, err)
		return err
	}

	for _, m := range messages {
		text.WriteString(fmt.Sprintf(
			"[%s - #%s] %s: %s\n",
			g.Name,
			ch.Name,
			m.Author.Username,
			m.Content,
		))
		log.Infof(ctx, "%v", m)
	}

	msg := slack.Message{Text: text.String()}
	if err := slackClient.Post(ctx, msg); err != nil {
		log.Errorf(ctx, "failed to post to slack: %v", err)
		return err
	}

	return nil
}
