package handler

import (
	"fmt"
	"net/http"

	"github.com/nownabe/discord2slack/pkg/delayedchecker"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type Handler struct {
	channels []string
}

func New(channels []string) *Handler {
	return &Handler{channels}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	for _, channelID := range h.channels {
		if err := delayedchecker.Call(ctx, channelID); err != nil {
			log.Errorf(ctx, "failed to call checker: %v", err)
		}
	}

	fmt.Fprintf(w, http.StatusText(http.StatusOK))
}
