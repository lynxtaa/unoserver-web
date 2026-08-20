// Package multipart provides functions for working with form-data/multipart
package multipart

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/lynxtaa/unoserver-web/internal/httperror"
)

var (
	errNoFile        = errors.New("no file part in request")
	errMalformedName = errors.New("malformed filename")
	errTooLarge      = errors.New("file too large")
)

// StoreSingleFile stores the first uploaded file from form-data/multipart to a new
// temp folder and returns it's path in file system along with a cleanup function
// removing that folder. Non-file fields are skipped. maxSizeBytes of 0 or less
// means unlimited.
func StoreSingleFile(
	r *http.Request,
	fieldName string,
	maxSizeBytes int64,
) (path string, cleanup func() error, err error) {
	reader, err := r.MultipartReader()
	if err != nil {
		return "", nil, httperror.New(err, "failed to create multipart reader", http.StatusBadRequest)
	}

	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			return "", nil, httperror.New(
				errNoFile,
				fmt.Sprintf("expected %q field", fieldName),
				http.StatusBadRequest,
			)
		}
		if err != nil {
			return "", nil, httperror.New(err, "failed to read part", http.StatusBadRequest)
		}

		if part.FileName() == "" {
			// Not a file, e.g. a plain text field
			_ = part.Close()
			continue
		}

		path, cleanup, err := storeFile(part, part.FileName(), maxSizeBytes)
		_ = part.Close()

		return path, cleanup, err
	}
}

func storeFile(
	src io.Reader,
	filename string,
	maxSizeBytes int64,
) (path string, cleanup func() error, err error) {
	filename = filepath.Base(filepath.Clean("/" + filename))
	if filename == "." || filename == string(filepath.Separator) {
		return "", nil, httperror.New(errMalformedName, "filename is malformed", http.StatusBadRequest)
	}

	folderPath, err := os.MkdirTemp("", "upload-*")
	if err != nil {
		return "", nil, fmt.Errorf("create temp folder: %w", err)
	}

	cleanup = func() error { return os.RemoveAll(folderPath) }

	path, err = copyToFolder(src, folderPath, filename, maxSizeBytes)
	if err != nil {
		_ = cleanup()
		return "", nil, err
	}

	return path, cleanup, nil
}

func copyToFolder(src io.Reader, folderPath, filename string, maxSizeBytes int64) (string, error) {
	targetPath := filepath.Join(folderPath, filename)

	//nolint:gosec // filename is base-named and folderPath is a fresh temp folder
	dst, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	if maxSizeBytes <= 0 {
		if _, err := io.Copy(dst, src); err != nil {
			return "", fmt.Errorf("save file: %w", err)
		}
		return targetPath, nil
	}

	written, err := io.CopyN(dst, src, maxSizeBytes+1)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("save file: %w", err)
	}

	if written > maxSizeBytes {
		return "", httperror.New(errTooLarge, "file too large", http.StatusRequestEntityTooLarge)
	}

	return targetPath, nil
}
