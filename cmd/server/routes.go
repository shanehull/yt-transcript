package main

import (
	"net/http"

	yt "github.com/shanehull/yt-transcript"
	"github.com/shanehull/yt-transcript/internal/cache"
	"github.com/shanehull/yt-transcript/internal/handlers"
	"github.com/shanehull/yt-transcript/internal/middleware"
)

func addRoutes(
	mux *http.ServeMux,
	client *yt.Client,
	transcriptCache *cache.Cache,
	baseURL string,
	allowedOrigin string,
	version string,
) {
	mux.HandleFunc(
		"/healthz",
		handlers.Healthz,
	)

	mux.Handle(
		"GET /{$}",
		middleware.CORS(
			handlers.Index(baseURL, version),
			allowedOrigin,
		),
	)

	mux.HandleFunc(
		"/favicon.ico",
		handlers.Favicon,
	)

	mux.Handle(
		"/{video_id}",
		middleware.CORS(
			handlers.Transcript(client, transcriptCache),
			allowedOrigin,
		),
	)
}
