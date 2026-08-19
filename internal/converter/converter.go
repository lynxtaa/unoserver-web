// Package converter provides interface for converting files
package converter

import (
	"context"
)

// ConvertOptions are options for converting
type ConvertOptions struct {
	Filter string
}

// Client defines the interface for document conversion
type Client interface {
	Convert(ctx context.Context, from, to string, opts ConvertOptions) error
}
