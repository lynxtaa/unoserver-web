package tests

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/lynxtaa/unoserver-web/internal/application"
	"github.com/lynxtaa/unoserver-web/internal/config"
	"github.com/lynxtaa/unoserver-web/internal/converter/unoserver"
	httpserver "github.com/lynxtaa/unoserver-web/internal/http"
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

	cfg := &config.Config{
		MaxWorkers:  8,
		MaxFileSize: 134217728,
	}

	uno := unoserver.New(unoserver.Options{
		MaxWorkers: cfg.MaxWorkers,
	})

	app := application.New(uno)
	srv := httpserver.NewServer(cfg, app)
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
