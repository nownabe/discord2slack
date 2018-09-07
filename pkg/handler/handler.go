package handler

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/nownabe/discord2slack/pkg/worker/guild"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

type Handler struct {
	guilds []string
}

func New(guilds []string) *Handler {
	return &Handler{guilds}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	for _, guildID := range h.guilds {
		v := url.Values{}
		v.Add(guild.GuildIDKey, guildID)
		t := taskqueue.NewPOSTTask("/worker/guild", v)
		if _, err := taskqueue.Add(ctx, t, ""); err != nil {
			log.Errorf(ctx, "failed to enqueue guild check task: %v", err)
		}
	}

	fmt.Fprintf(w, http.StatusText(http.StatusOK))
}
