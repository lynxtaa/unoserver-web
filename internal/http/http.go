// Package http handles server creation
package http

import (
	"net/http"
	"strings"

	"github.com/lynxtaa/unoserver-web/internal/application"
	"github.com/lynxtaa/unoserver-web/internal/config"
	"github.com/lynxtaa/unoserver-web/internal/cors"
	"github.com/lynxtaa/unoserver-web/internal/httplog"
	"github.com/lynxtaa/unoserver-web/internal/reqid"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

const documentationPath = "/documentation"

// Server represents the HTTP server.
type Server struct {
	mux         *http.ServeMux
	application *application.App
	cfg         *config.Config
}

// NewServer creates a new HTTP server with all routes configured.
func NewServer(cfg *config.Config, application *application.App) *Server {
	s := &Server{
		mux:         http.NewServeMux(),
		application: application,
		cfg:         cfg,
	}

	basePath := strings.TrimSuffix(cfg.BasePath, "/")

	s.mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, basePath+documentationPath+"/index.html", http.StatusFound)
	})

	// Aliases for URLs served by the previous Fastify implementation
	s.mux.HandleFunc("GET /documentation/json", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, basePath+documentationPath+"/doc.json", http.StatusFound)
	})

	s.mux.HandleFunc("GET /documentation/static/{file...}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, basePath+documentationPath+"/"+r.PathValue("file"), http.StatusFound)
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
