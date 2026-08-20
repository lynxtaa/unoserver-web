// Package server handles HTTP server creation
package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/lynxtaa/unoserver-web/internal/config"
	"github.com/lynxtaa/unoserver-web/internal/converter"
	"github.com/lynxtaa/unoserver-web/internal/cors"
	"github.com/lynxtaa/unoserver-web/internal/httplog"
	"github.com/lynxtaa/unoserver-web/internal/reqid"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

const documentationPath = "/documentation"

// Converter converts a file on disk into another format
type Converter interface {
	Convert(ctx context.Context, from, to string, opts converter.ConvertOptions) error
}

// Server represents the HTTP server.
type Server struct {
	mux       *http.ServeMux
	converter Converter
	cfg       *config.Config
}

// NewServer creates a new HTTP server with all routes configured.
func NewServer(cfg *config.Config, converter Converter) *Server {
	s := &Server{
		mux:       http.NewServeMux(),
		converter: converter,
		cfg:       cfg,
	}

	basePath := strings.TrimSuffix(cfg.BasePath, "/")

	s.mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, basePath+documentationPath+"/index.html", http.StatusFound)
	})

	// Aliases for URLs served by the previous Fastify implementation
	s.mux.HandleFunc("GET /documentation/json", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, basePath+documentationPath+"/doc.json", http.StatusFound)
	})

	s.mux.HandleFunc("GET /documentation/static/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, basePath+documentationPath+"/index.html", http.StatusFound)
	})

	s.mux.HandleFunc("GET /documentation/{any...}", httpSwagger.WrapHandler)

	s.mux.HandleFunc("POST /convert/{format}", s.handleUpload)

	return s
}

// Handler returns the HTTP Handler with middleware applied.
func (s *Server) Handler() http.Handler {
	handler := http.Handler(s.mux)
	handler = httplog.Middleware(handler)
	handler = reqid.Middleware(s.cfg.RequestIDHeader)(handler)
	handler = cors.Middleware(handler)

	return handler
}
