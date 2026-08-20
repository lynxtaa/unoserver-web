// Package httperror provides HTTP error handling utilities
package httperror

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// ErrorResponse is a JSON body returned for any failed request
type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error"`
	Message    string `json:"message"`
}

// HTTPError represents an HTTP error with status code and response message
type HTTPError struct {
	Err      error
	Status   int
	Response string
}

func (e *HTTPError) Error() string {
	if e.Err == nil {
		return e.Response
	}
	return e.Err.Error()
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

// WriteTo writes the error response to the http.ResponseWriter
func (e *HTTPError) WriteTo(ctx context.Context, w http.ResponseWriter) {
	writeJSON(ctx, w, e.Status, e.Response)
}

// New creates a new HTTPError with the given error, response message and status code
func New(err error, res string, status int) *HTTPError {
	return &HTTPError{
		Err:      err,
		Response: res,
		Status:   status,
	}
}

// RespondWithError writes an error response to the http.ResponseWriter
func RespondWithError(ctx context.Context, err error, w http.ResponseWriter) {
	if httpErr, ok := errors.AsType[*HTTPError](err); ok {
		if httpErr.Status < http.StatusInternalServerError {
			slog.WarnContext(ctx, "Client error", "error", httpErr.Response)
		} else {
			slog.ErrorContext(ctx, "Server error", "error", httpErr)
		}
		httpErr.WriteTo(ctx, w)
		return
	}

	slog.ErrorContext(ctx, "Server error", "error", err)
	writeJSON(ctx, w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
}

func writeJSON(ctx context.Context, w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)

	body := ErrorResponse{
		StatusCode: status,
		Error:      http.StatusText(status),
		Message:    message,
	}

	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.WarnContext(ctx, "write error response failed", "error", err)
	}
}
