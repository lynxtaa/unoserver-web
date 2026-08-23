// Package server handles HTTP server creation
package server

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/lynxtaa/unoserver-web/internal/config"
	"github.com/lynxtaa/unoserver-web/internal/converter"
	"github.com/lynxtaa/unoserver-web/internal/cors"
	"github.com/lynxtaa/unoserver-web/internal/httplog"
	"github.com/lynxtaa/unoserver-web/internal/reqid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// Converter converts a file on disk into another format
type Converter interface {
	Convert(ctx context.Context, from, to string, opts converter.ConvertOptions) error
}

// Server represents the HTTP server.
type Server struct {
	mux       *http.ServeMux
	converter Converter
	cfg       *config.Config
	basePath  string
}

// NewServer creates a new HTTP server with all routes configured.
func NewServer(cfg *config.Config, converter Converter) *Server {
	s := &Server{
		mux:       http.NewServeMux(),
		converter: converter,
		cfg:       cfg,
		basePath:  strings.TrimSuffix(cfg.BasePath, "/"),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, s.basePath+"/documentation/index.html", http.StatusFound)
	})

	mux.HandleFunc("GET /documentation/{any...}", httpSwagger.WrapHandler)

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			slog.WarnContext(r.Context(), "responding to /health", "error", err)
		}
	})

	mux.Handle("GET /metrics", promhttp.Handler())

	mux.HandleFunc("POST /convert/{format}", s.handleUpload)

	if s.basePath != "" {
		s.mux.Handle(s.basePath+"/", http.StripPrefix(s.basePath, mux))
	} else {
		s.mux = mux
	}

	return s
}

// Handler returns the HTTP Handler with middleware applied.
func (s *Server) Handler() http.Handler {
	handler := http.Handler(s.mux)
	handler = httplog.Middleware(handler)
	handler = reqid.Middleware(s.cfg.RequestIDHeader)(handler)
	handler = cors.Middleware(handler)

	// Bypass all middleware for health checks and metrics
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, s.basePath)
		if path == "/health" || path == "/metrics" {
			s.mux.ServeHTTP(w, r)
			return
		}
		handler.ServeHTTP(w, r)
	})
}
