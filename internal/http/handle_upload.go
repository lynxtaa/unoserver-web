package http

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
// @Failure 400 {object} map[string]string
// @Failure 408 {object} map[string]string "Conversion timeout"
// @Failure 500 {object} map[string]string
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

	// nolint gosec
	file, err := os.Open(targetPath)
	if err != nil {
		httperror.RespondWithError(ctx, err, w)
		return
	}
	defer os.RemoveAll(filepath.Dir(targetPath))
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		httperror.RespondWithError(ctx, err, w)
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(targetPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)

	filename, _ := strings.CutSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
	disposition := mime.FormatMediaType("attachment", map[string]string{
		"filename": filename + filepath.Ext(targetPath),
	})
	if disposition != "" {
		w.Header().Set("Content-Disposition", disposition)
	}
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))

	w.WriteHeader(http.StatusOK)

	if _, err := io.Copy(w, file); err != nil {
		slog.WarnContext(ctx, "write response failed", "error", err)
		return
	}
}
