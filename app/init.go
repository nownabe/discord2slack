package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/nownabe/discord2slack/pkg/handler"
	"github.com/nownabe/discord2slack/pkg/worker/channel"
	"github.com/nownabe/discord2slack/pkg/worker/guild"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func init() {
	discordToken := os.Getenv("DISCORD_TOKEN")
	slackWebhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	guilds := strings.Split(os.Getenv("DISCORD_GUILD_IDS"), ",")

	h := handler.New(guilds)
	guildHandler := guild.New(discordToken)
	channelHandler := channel.New(discordToken, slackWebhookURL)

	mux := http.NewServeMux()
	mux.HandleFunc("/_ah/health", healthCheckHandler)
	mux.Handle("/check", h)
	mux.Handle("/worker/guild", guildHandler)
	mux.Handle("/worker/channel", channelHandler)
	mux.Handle("/", http.NotFoundHandler())
	http.Handle("/", mux)
}
