package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/nownabe/discord2slack/pkg/delayedchecker"
	"github.com/nownabe/discord2slack/pkg/handler"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func init() {
	delayedchecker.DiscordToken = os.Getenv("DISCORD_TOKEN")
	delayedchecker.SlackWebhookURL = os.Getenv("SLACK_WEBHOOK_URL")

	channels := strings.Split(os.Getenv("DISCORD_CHANNEL_IDS"), ",")

	h := handler.New(channels)

	mux := http.NewServeMux()
	mux.HandleFunc("/_ah/health", healthCheckHandler)
	mux.Handle("/check", h)
	mux.Handle("/", http.NotFoundHandler())
	http.Handle("/", mux)
}
