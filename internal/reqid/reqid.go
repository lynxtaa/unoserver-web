// Package reqid attaches a random request ID to the request context and logs
// it via slog.
package reqid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"

	"github.com/lynxtaa/unoserver-web/internal/config"
)

type ctxKey struct{}

// New returns a random request ID.
func New() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// WithID returns a copy of ctx carrying the request ID.
func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// FromContext returns the request ID stored in ctx, or "" if absent.
func FromContext(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

// Middleware generates a request ID, stores it in the request context, and
// echoes it in the X-Request-Id response header.
func Middleware(cfg *config.Config) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := New()
			if cfg.RequestIDHeader != "" {
				if reqID := r.Header.Get(cfg.RequestIDHeader); reqID != "" {
					id = reqID
				}
			}
			w.Header().Set("X-Request-Id", id)
			next.ServeHTTP(w, r.WithContext(WithID(r.Context(), id)))
		})
	}
}

// Handler wraps h so records logged with a context carrying a request ID
// include it as the "reqId" attribute.
func Handler(h slog.Handler) slog.Handler {
	return &slogHandler{h}
}

type slogHandler struct{ slog.Handler }

func (h *slogHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := FromContext(ctx); id != "" {
		r.AddAttrs(slog.String("reqId", id))
	}
	return h.Handler.Handle(ctx, r)
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &slogHandler{h.Handler.WithAttrs(attrs)}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	return &slogHandler{h.Handler.WithGroup(name)}
}
