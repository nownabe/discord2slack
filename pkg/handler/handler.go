package handler

import (
	"fmt"
	"net/http"

	"github.com/nownabe/discord2slack/pkg/delayedchecker"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
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
		if err := delayedchecker.Call(ctx, guildID); err != nil {
			log.Errorf(ctx, "failed to call checker: %v", err)
		}
	}

	fmt.Fprintf(w, http.StatusText(http.StatusOK))
}
