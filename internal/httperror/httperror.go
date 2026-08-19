// Package httperror provides HTTP error handling utilities
package httperror

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// HTTPError represents an HTTP error with status code and response message
type HTTPError struct {
	Err      error
	Status   int
	Response string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%v", e.Err)
}

func (e *HTTPError) Unwrap() error {
	return e.Err
}

// WriteTo writes the error response to the http.ResponseWriter
func (e *HTTPError) WriteTo(w http.ResponseWriter) {
	http.Error(w, e.Response, e.Status)
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
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		if httpErr.Status < 500 {
			slog.WarnContext(ctx, "Client error", "error", httpErr.Response)
		} else {
			slog.ErrorContext(ctx, "Server error", "error", httpErr)
		}
		httpErr.WriteTo(w)
	} else {
		slog.ErrorContext(ctx, "Server error", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
