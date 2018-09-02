package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/nownabe/discord2slack/pkg/checker"
	"github.com/nownabe/discord2slack/pkg/slack"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func init() {
	token := os.Getenv("DISCORD_TOKEN")
	channels := os.Getenv("DISCORD_CHANNEL_IDS")

	slack, err := slack.New(os.Getenv("SLACK_WEBHOOK_URL"))
	if err != nil {
		panic(err)
	}

	checkHandler, err := checker.New(token, strings.Split(channels, ","), slack)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/_ah/health", healthCheckHandler)
	mux.Handle("/check", checkHandler)
	mux.Handle("/", http.NotFoundHandler())
	http.Handle("/", mux)
}
