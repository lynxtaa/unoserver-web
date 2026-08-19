// Package main is the entry point for the unoserver-web application
//
// @title unoserver-web
// @version 0.1.0
// @BasePath /
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
	_ "github.com/lynxtaa/unoserver-web/docs"
	"github.com/lynxtaa/unoserver-web/internal/application"
	"github.com/lynxtaa/unoserver-web/internal/config"
	"github.com/lynxtaa/unoserver-web/internal/converter/unoserver"
	httpserver "github.com/lynxtaa/unoserver-web/internal/http"
	"github.com/lynxtaa/unoserver-web/internal/reqid"
)

const (
	gracefulShutdownTimeout = 15 * time.Second
	readHeaderTimeout       = 10 * time.Second
)

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load(ctx)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	var handler slog.Handler
	if cfg.PrettyLogs {
		handler = tint.NewHandler(os.Stdout, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stderr, nil)
	}

	logger := slog.New(reqid.Handler(handler))
	slog.SetDefault(logger)

	unoserver := unoserver.New(unoserver.Options{
		MaxWorkers:        cfg.MaxWorkers,
		ConversionRetries: &cfg.ConversionRetries,
	})

	app := application.New(unoserver)

	httpServer := httpserver.NewServer(cfg, app)

	addr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           httpServer.Handler(),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, "HTTP server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		slog.ErrorContext(ctx, "HTTP server", "error", err)
		return err
	case <-ctx.Done():
		slog.InfoContext(ctx, "Shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		unoserver.StopServer(ctx)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "server: %v\n", err)
		os.Exit(1)
	}
}
