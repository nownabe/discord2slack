package checker

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/nownabe/discord2slack/pkg/slack"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type Handler struct {
	discord  *discordgo.Session
	channels []string
	slack    *slack.Client
}

func New(token string, channels []string, s *slack.Client) (*Handler, error) {
	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	return &Handler{
		discord:  discord,
		channels: channels,
		slack:    s,
	}, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)
	h.discord.Client = client

	text := new(bytes.Buffer)

	for _, c := range h.channels {
		ch, err := h.discord.Channel(c)
		if err != nil {
			log.Errorf(ctx, "faild to get discord channel: %v", err)
			continue
		}

		g, err := h.discord.Guild(ch.GuildID)
		if err != nil {
			log.Errorf(ctx, "failed to get discord guild: %v", err)
			continue
		}

		messages, err := h.discord.ChannelMessages(c, 100, "", "", "")
		if err != nil {
			log.Errorf(ctx, "faild to get discord messages at %s in %s: %v", ch.Name, g.Name, err)
			continue
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
	}

	msg := slack.Message{Text: text.String()}
	if err := h.slack.Post(ctx, msg); err != nil {
		log.Errorf(ctx, "failed to post to slack: %v", err)
		http.Error(w, "failed", http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "checked")
}
