package http

import (
	"context"
	"errors"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/lynxtaa/unoserver-web/internal/converter"
	"github.com/lynxtaa/unoserver-web/internal/httperror"
	"github.com/lynxtaa/unoserver-web/internal/multipart"
)

// handleUpload godoc
// @Summary Convert file to specified format
// @Description Upload a file and convert it to the specified format
// @Tags conversion
// @Accept multipart/form-data
// @Produce application/octet-stream
// @Param format path string true "Target format (pdf, docx, odt, etc)"
// @Param file formData file true "File to convert"
// @Param filter query string false "Export filter options"
// @Success 200 {file} binary "Converted file"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 408 {object} httperror.ErrorResponse "Conversion timeout"
// @Failure 413 {object} httperror.ErrorResponse "File too large"
// @Failure 500 {object} httperror.ErrorResponse
// @Router /convert/{format} [post]
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	format := r.PathValue("format")
	filter := r.URL.Query().Get("filter")

	srcPath, err := multipart.StoreSingleFile(r, "file", s.cfg.MaxFileSize)
	if err != nil {
		httperror.RespondWithError(ctx, err, w)
		return
	}

	// Cleaned up regardless of how the request ends, the converted file lands here too
	defer func() {
		if err := os.RemoveAll(filepath.Dir(srcPath)); err != nil {
			slog.WarnContext(ctx, "removing temp folder failed", "error", err)
		}
	}()

	slog.InfoContext(ctx, "file uploaded", "path", srcPath)

	targetPath, err := s.application.ConvertFile.Handle(ctx, srcPath, format, converter.ConvertOptions{
		Filter: filter,
	})
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			httperror.RespondWithError(
				ctx,
				httperror.New(err, "conversion timeout", http.StatusRequestTimeout),
				w,
			)
			return
		}

		httperror.RespondWithError(ctx, err, w)
		return
	}

	//nolint:gosec // targetPath is derived from a temp folder created by this process
	file, err := os.Open(targetPath)
	if err != nil {
		httperror.RespondWithError(ctx, err, w)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		httperror.RespondWithError(ctx, err, w)
		return
	}

	w.Header().Set("Content-Type", contentType(filepath.Ext(targetPath)))

	filename, _ := strings.CutSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
	disposition := mime.FormatMediaType("attachment", map[string]string{
		"filename": filename + filepath.Ext(targetPath),
	})
	if disposition != "" {
		w.Header().Set("Content-Disposition", disposition)
	}

	// ServeContent sets Content-Length and handles range requests
	http.ServeContent(w, r, targetPath, stat.ModTime(), file)
}
