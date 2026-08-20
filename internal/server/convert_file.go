package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/lynxtaa/unoserver-web/internal/converter"
	"github.com/lynxtaa/unoserver-web/internal/httperror"
)

// ErrInvalidExtension is returned when a file has no extension
var ErrInvalidExtension = errors.New("extension is empty")

// convertFile converts srcPath into the given format, placing the result next to
// the source file, and returns its path
func (s *Server) convertFile(
	ctx context.Context,
	srcPath string,
	format string,
	opts converter.ConvertOptions,
) (targetPath string, err error) {
	ext := strings.ToLower(filepath.Ext(srcPath))
	if ext == "" {
		return "", httperror.New(
			ErrInvalidExtension,
			"can't detect extension for incoming file",
			http.StatusBadRequest,
		)
	}

	pathWithoutExtension := strings.TrimSuffix(srcPath, ext)

	if sameSrcAndTargetFormat := "."+format == ext; sameSrcAndTargetFormat {
		pathWithoutExtension += "-1"
	}

	targetPath = pathWithoutExtension + "." + format

	if err := s.converter.Convert(ctx, srcPath, targetPath, opts); err != nil {
		return "", fmt.Errorf("conversion failed: %w", err)
	}

	return targetPath, nil
}
