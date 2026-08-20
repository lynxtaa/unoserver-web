// Package httplog provides HTTP middleware that logs incoming requests via slog.
package httplog

import (
	"log/slog"
	"net/http"
	"time"
)

// Middleware logs each incoming request and its outcome. Headers are deliberately
// not logged, they carry credentials.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		slog.InfoContext(r.Context(), "incoming request",
			"method", r.Method,
			"path", r.URL.Path,
			"remoteAddr", r.RemoteAddr,
		)

		recorder := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		slog.InfoContext(r.Context(), "request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.Status(),
			"responseTime", time.Since(now).Milliseconds(),
		)
	})
}

// statusRecorder remembers the status code written to the response
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

// Unwrap exposes the wrapped writer to http.ResponseController
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// Status returns the written status code, defaulting to 200
func (r *statusRecorder) Status() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}
