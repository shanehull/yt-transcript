// Package middleware provides HTTP middleware for the yt-transcript server.
package middleware

import (
	"net/http"
)

// CSP sets a Content-Security-Policy header on every response.
func CSP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'none'; "+
				"style-src 'self' 'unsafe-inline'; "+
				"img-src 'self' data:; "+
				"base-uri 'self'; "+
				"frame-ancestors 'none'; "+
				"form-action 'none'; "+
				"object-src 'none';")
		next.ServeHTTP(w, r)
	})
}
