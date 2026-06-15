// Package middleware provides HTTP middleware for the yt-transcript server.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// CtxKey is the context key for cache status in request context.
type CtxKey struct{}

// CacheStatus tracks whether a request was a cache hit or miss.
type CacheStatus struct {
	Hit bool
	Set bool
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// Logging logs each request with method, path, status, cache, and duration.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		cs := &CacheStatus{}
		ctx := context.WithValue(r.Context(), CtxKey{}, cs)
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r.WithContext(ctx))

		args := []any{
			"method", r.Method,
			"path", r.URL.Path,
			"status", sw.status,
		}
		if cs.Set {
			if cs.Hit {
				args = append(args, "cache", "HIT")
			} else {
				args = append(args, "cache", "MISS")
			}
		}
		args = append(args, "duration_ms", time.Since(start).Milliseconds())
		slog.Info("request", args...)
	})
}
