// Package cors provides CORS middleware
package cors

import "net/http"

const maxAgeSeconds = "3600"

const allowedMethods = "GET, HEAD, PUT, PATCH, POST, DELETE"

// Middleware wraps any http.Handler to apply generic CORS headers
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")

		isPreflight := r.Method == http.MethodOptions &&
			r.Header.Get("Access-Control-Request-Method") != ""

		if !isPreflight {
			// A plain OPTIONS request is routed as usual
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Access-Control-Allow-Methods", allowedMethods)
		// Reflect requested headers, so custom ones like REQUEST_ID_HEADER pass preflight
		if headers := r.Header.Get("Access-Control-Request-Headers"); headers != "" {
			w.Header().Set("Access-Control-Allow-Headers", headers)
			w.Header().Add("Vary", "Access-Control-Request-Headers")
		}
		w.Header().Set("Access-Control-Max-Age", maxAgeSeconds)
		w.Header().Add("Vary", "Access-Control-Request-Method")

		w.WriteHeader(http.StatusNoContent)
	})
}
