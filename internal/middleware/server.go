package middleware

import (
	"net/http"
)

// ServerHeader sets the Server and X-yt-transcript-Version response headers.
func ServerHeader(version string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "yt-transcript")
		w.Header().Set("X-yt-transcript-Version", version)
		next.ServeHTTP(w, r)
	})
}
