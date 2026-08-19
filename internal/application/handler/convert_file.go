// Package handler provides application handlers
package handler

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lynxtaa/unoserver-web/internal/converter"
)

// ConvertFileHandler handles file conversion operations
type ConvertFileHandler struct {
	converter converter.Client
}

// NewConvertFileHandler creates a new ConvertFileHandler with the given converter implementation
func NewConvertFileHandler(converter converter.Client) *ConvertFileHandler {
	return &ConvertFileHandler{
		converter: converter,
	}
}

// ErrInvalidExtension is returned when a file has no extension
var ErrInvalidExtension = errors.New("extension is empty")

// Handle converts a file from srcPath to the specified format and returns the target path
func (c *ConvertFileHandler) Handle(
	ctx context.Context,
	srcPath string,
	format string,
	opts converter.ConvertOptions,
) (targetPath string, err error) {
	ext := strings.ToLower(filepath.Ext(srcPath))
	if ext == "" {
		return "", ErrInvalidExtension
	}
	pathWithoutExtension := strings.TrimSuffix(srcPath, ext)

	if sameSrcAndTargetFormat := "."+format == ext; sameSrcAndTargetFormat {
		pathWithoutExtension += "-1"
	}

	targetPath = pathWithoutExtension + "." + format

	if err := c.converter.Convert(ctx, srcPath, targetPath, opts); err != nil {
		return "", fmt.Errorf("conversion failed: %w", err)
	}

	return targetPath, nil
}
