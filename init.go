package main

import (
	"fmt"
	"net/http"
)

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func init() {
	mux := http.NewServeMux()
	mux.HandleFunc("/_ah/health", healthCheckHandler)
	mux.Handle("/", http.NotFoundHandler())
	http.Handle("/", mux)
}
