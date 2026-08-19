// Package multipart provides functions for working with form-data/multipart
package multipart

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/lynxtaa/unoserver-web/internal/httperror"
)

// StoreSingleFile stores singe file from form-data/multipart to a new temp folder
// and returns it's path in file system
func StoreSingleFile(
	r *http.Request,
	fieldName string,
	maxSizeBytes int,
) (string, error) {
	reader, err := r.MultipartReader()
	if err != nil {
		return "", httperror.New(err, "failed to create multipart reader", http.StatusBadRequest)
	}

	part, err := reader.NextPart()
	if err != nil {
		return "", httperror.New(err, "failed to read part", http.StatusBadRequest)
	}
	defer part.Close()

	if part.FormName() != fieldName {
		return "", httperror.New(nil, fmt.Sprintf("expected %q field", fieldName), http.StatusBadRequest)
	}

	filename := part.FileName()
	if filename == "" {
		return "", httperror.New(nil, "filename is required", http.StatusBadRequest)
	}

	folderPath, err := os.MkdirTemp("", "upload-*")
	if err != nil {
		return "", fmt.Errorf("create temp folder: %w", err)
	}

	targetPath := filepath.Join(folderPath, filename)
	targetPath = filepath.Clean(targetPath)
	if !strings.HasPrefix(targetPath, folderPath) {
		return "", httperror.New(nil, "filename is malformed", http.StatusBadRequest)
	}

	dst, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	written, err := io.CopyN(dst, part, int64(maxSizeBytes)+1)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("save file: %w", err)
	}

	if written > int64(maxSizeBytes) {
		_ = os.RemoveAll(folderPath)
		return "", httperror.New(nil, "file too large", http.StatusRequestEntityTooLarge)
	}

	return targetPath, nil
}
