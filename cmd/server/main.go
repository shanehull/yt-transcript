// Command yt-transcript-server serves YouTube transcripts over HTTP.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	yt "github.com/shanehull/yt-transcript"
	"github.com/shanehull/yt-transcript/internal/cache"
	"github.com/shanehull/yt-transcript/internal/middleware"
)

var (
	logger  *slog.Logger
	version = "dev"
)

func init() {
	logger = slog.New(
		slog.NewJSONHandler(os.Stdout, nil),
	).With(slog.String("version", version))
	slog.SetDefault(logger)
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Getenv); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, getenv func(string) string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()
	host := "127.0.0.1"
	port := "8080"
	allowedOrigin := "*"
	baseURL := ""

	if v := getenv("SERVER_HOST"); v != "" {
		host = v
	}
	if v := getenv("PORT"); v != "" {
		port = v
	}
	if v := getenv("ALLOWED_ORIGIN"); v != "" {
		allowedOrigin = v
	}
	baseURL = getenv("BASE_URL")
	if baseURL == "" {
		baseURL = fmt.Sprintf("http://%s:%s", host, port)
	}

	client := yt.NewClient()
	transcriptCache := cache.New(1 * time.Hour)

	srv := NewServer(client, transcriptCache, baseURL, allowedOrigin, version)

	addr := fmt.Sprintf("%s:%s", host, port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      srv,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	logger.Info("server listening", "addr", addr)

	<-ctx.Done()
	logger.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("forced shutdown", "error", err)
	}
	logger.Info("server stopped")
	return nil
}

// NewServer builds the application http.Handler with all routes and middleware.
func NewServer(
	client *yt.Client,
	transcriptCache *cache.Cache,
	baseURL string,
	allowedOrigin string,
	version string,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, client, transcriptCache, baseURL, allowedOrigin, version)
	var handler http.Handler = mux
	handler = middleware.CSP(handler)
	handler = middleware.ServerHeader(version, handler)
	handler = middleware.Logging(handler)
	return handler
}
