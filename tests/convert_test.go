package tests

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/lynxtaa/unoserver-web/internal/config"
	"github.com/lynxtaa/unoserver-web/internal/converter/unoserver"
	"github.com/lynxtaa/unoserver-web/internal/httperror"
	"github.com/lynxtaa/unoserver-web/internal/server"
)

//go:embed fixtures/1.rtf
var rtfFile []byte

type testResponse struct {
	statusCode int
	header     http.Header
	body       []byte
}

func startTestServer(t *testing.T) (*httptest.Server, func()) {
	t.Helper()

	return startTestServerWithConfig(t, &config.Config{
		MaxWorkers:        8,
		ConversionRetries: 3,
		LogLevel:          slog.LevelError,
	})
}

func startTestServerWithConfig(t *testing.T, cfg *config.Config) (*httptest.Server, func()) {
	t.Helper()

	uno := unoserver.New(unoserver.Options{
		MaxWorkers:        cfg.MaxWorkers,
		ConversionRetries: cfg.ConversionRetries,
	})

	srv := server.NewServer(cfg, uno)
	ts := httptest.NewServer(srv.Handler())

	cleanup := func() {
		ts.Close()
		uno.StopServer(context.Background())
	}

	return ts, cleanup
}

func postFile(client *http.Client, url, filename string, data []byte) (*testResponse, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}

	if _, err := part.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &testResponse{
		statusCode: resp.StatusCode,
		header:     resp.Header,
		body:       respBody,
	}, nil
}

func TestConvertDocx(t *testing.T) {
	ts, cleanup := startTestServer(t)
	defer cleanup()

	res, err := postFile(ts.Client(), ts.URL+"/convert/docx", "1.rtf", rtfFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.statusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.statusCode, string(res.body))
	}

	contentType := res.header.Get("Content-Type")
	if contentType != "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		t.Errorf("expected Content-Type application/vnd.openxmlformats-officedocument.wordprocessingml.document, got %s", contentType)
	}

	contentDisposition := res.header.Get("Content-Disposition")
	if contentDisposition != "attachment; filename=1.docx" {
		t.Errorf("expected Content-Disposition attachment; filename=1.docx, got %s", contentDisposition)
	}

	res2, err := postFile(ts.Client(), ts.URL+"/convert/fodt", "1.docx", res.body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res2.statusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res2.statusCode, string(res2.body))
	}

	contentDisposition2 := res2.header.Get("Content-Disposition")
	if contentDisposition2 != "attachment; filename=1.fodt" {
		t.Errorf("expected Content-Disposition attachment; filename=1.fodt, got %s", contentDisposition2)
	}

	if !strings.Contains(string(res2.body), "Hello World!") {
		t.Errorf("expected body to contain 'Hello World!', got %s", string(res2.body))
	}
}

func TestConvertRtf(t *testing.T) {
	ts, cleanup := startTestServer(t)
	defer cleanup()

	res, err := postFile(ts.Client(), ts.URL+"/convert/rtf", "1.rtf", rtfFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.statusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.statusCode, string(res.body))
	}

	contentType := res.header.Get("Content-Type")
	if contentType != "application/rtf" {
		t.Errorf("expected Content-Type application/rtf, got %s", contentType)
	}

	contentDisposition := res.header.Get("Content-Disposition")
	if contentDisposition != "attachment; filename=1.rtf" {
		t.Errorf("expected Content-Disposition attachment; filename=1.rtf, got %s", contentDisposition)
	}

	if !strings.Contains(string(res.body), "Hello World!") {
		t.Errorf("expected body to contain 'Hello World!', got %s", string(res.body))
	}
}

// bigRtf builds an RTF document larger than 1 MiB
func bigRtf() []byte {
	const paragraph = `\pard\sa200\sl276\slmult1\f0\fs22\lang9 Hello World!\par` + "\n"

	var buf bytes.Buffer
	buf.WriteString(`{\rtf1\ansi\deff0{\fonttbl{\f0\fnil\fcharset0 Calibri;}}` + "\n")
	for range 30_000 {
		buf.WriteString(paragraph)
	}
	buf.WriteString("}")

	return buf.Bytes()
}

func TestConvertBiggerThan1MiB(t *testing.T) {
	ts, cleanup := startTestServer(t)
	defer cleanup()

	data := bigRtf()
	if len(data) <= 1024*1024 {
		t.Fatalf("expected fixture bigger than 1 MiB, got %d bytes", len(data))
	}

	res, err := postFile(ts.Client(), ts.URL+"/convert/rtf", "big.rtf", data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.statusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", res.statusCode, string(res.body))
	}
}

func TestFileTooLarge(t *testing.T) {
	ts, cleanup := startTestServerWithConfig(t, &config.Config{
		MaxWorkers:  8,
		MaxFileSize: 1024,
		LogLevel:    slog.LevelError,
	})
	defer cleanup()

	res, err := postFile(ts.Client(), ts.URL+"/convert/pdf", "big.rtf", bigRtf())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.statusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d: %s", res.statusCode, string(res.body))
	}
}

func TestMissingFileReturnsJSONError(t *testing.T) {
	ts, cleanup := startTestServer(t)
	defer cleanup()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("filter", "writer_pdf_Export"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, ts.URL+"/convert/pdf", &body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := ts.Client().Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}

	if contentType := res.Header.Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Errorf("expected JSON content type, got %s", contentType)
	}

	var errRes httperror.ErrorResponse
	if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
		t.Fatalf("decoding error response: %v", err)
	}

	if errRes.StatusCode != http.StatusBadRequest || errRes.Message != `expected "file" field` {
		t.Errorf("unexpected error response: %+v", errRes)
	}
}

func TestParallelConversion(t *testing.T) {
	ts, cleanup := startTestServer(t)
	defer cleanup()

	const count = 64
	var wg sync.WaitGroup
	errCh := make(chan error, count)

	for range count {
		wg.Go(func() {
			res, err := postFile(ts.Client(), ts.URL+"/convert/pdf", "1.rtf", rtfFile)
			if err != nil {
				errCh <- err
				return
			}
			if res.statusCode != http.StatusOK {
				errCh <- fmt.Errorf("unexpected status %d: %s", res.statusCode, string(res.body))
				return
			}
		})
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
}
