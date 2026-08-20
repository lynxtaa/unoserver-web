// Package reqid attaches a random request ID to the request context and logs
// it via slog.
package reqid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
)

// DefaultLogLabel is the slog attribute name used when no label is configured
const DefaultLogLabel = "reqId"

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
// echoes it in the X-Request-Id response header. When headerName is not empty,
// an incoming header of that name is used instead of a generated ID.
func Middleware(headerName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := New()
			if headerName != "" {
				if reqID := r.Header.Get(headerName); reqID != "" {
					id = reqID
				}
			}
			w.Header().Set("X-Request-Id", id)
			next.ServeHTTP(w, r.WithContext(WithID(r.Context(), id)))
		})
	}
}

// Handler wraps h so records logged with a context carrying a request ID
// include it as an attribute named logLabel (DefaultLogLabel when empty).
func Handler(h slog.Handler, logLabel string) slog.Handler {
	if logLabel == "" {
		logLabel = DefaultLogLabel
	}
	return &slogHandler{Handler: h, logLabel: logLabel}
}

type slogHandler struct {
	slog.Handler
	logLabel string
}

func (h *slogHandler) Handle(ctx context.Context, r slog.Record) error {
	if id := FromContext(ctx); id != "" {
		r.AddAttrs(slog.String(h.logLabel, id))
	}
	return h.Handler.Handle(ctx, r)
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &slogHandler{Handler: h.Handler.WithAttrs(attrs), logLabel: h.logLabel}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	return &slogHandler{Handler: h.Handler.WithGroup(name), logLabel: h.logLabel}
}
