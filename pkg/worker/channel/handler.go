package channel

import (
	"fmt"
	"net/http"

	"github.com/nownabe/discord2slack/pkg/entity"
	"github.com/nownabe/discord2slack/pkg/slack"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

const (
	GuildIDKey     = "guildID"
	GuildNameKey   = "guildName"
	ChannelIDKey   = "channelID"
	ChannelNameKey = "channelName"
)

type Handler struct {
	discordToken string
	slack        *slack.Client
}

func New(discordToken, slackWebhookURL string) *Handler {
	return &Handler{discordToken, slack.New(slackWebhookURL)}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	guild := &entity.Guild{ID: r.FormValue(GuildIDKey), Name: r.FormValue(GuildNameKey)}
	channel := &entity.Channel{ID: r.FormValue(ChannelIDKey), Name: r.FormValue(ChannelNameKey)}

	chk, err := h.newChecker(ctx, guild, channel)
	if err != nil {
		log.Errorf(ctx, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := chk.check(ctx); err != nil {
		log.Errorf(ctx, err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "ok")
}
