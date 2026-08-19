// Package unoserver provides Python's `unoserver`
// for handling LibreOffice conversions
package unoserver

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/lynxtaa/unoserver-web/internal/converter"
	"github.com/lynxtaa/unoserver-web/internal/process"
	"github.com/lynxtaa/unoserver-web/internal/retry"
)

const (
	startWaitTime    = 5 * time.Second
	shutdownWaitTime = 10 * time.Second
)

// Unoserver contains everything related to `unoserver`
type Unoserver struct {
	semaphore         chan struct{}
	timeout           time.Duration
	conversionRetries int
	port              int
	mu                sync.Mutex
	process           *os.Process
}

// Options are unoserver options
type Options struct {
	MaxWorkers        int
	Timeout           time.Duration
	Port              *int
	ConversionRetries *int
}

// New returns new unoserver
func New(opts Options) *Unoserver {
	u := &Unoserver{
		semaphore:         make(chan struct{}, opts.MaxWorkers),
		timeout:           1 * time.Minute,
		port:              12345,
		conversionRetries: 3,
	}

	if opts.ConversionRetries != nil {
		u.conversionRetries = *opts.ConversionRetries
	}
	if opts.Port != nil {
		u.port = *opts.Port
	}
	if opts.Timeout != 0 {
		u.timeout = opts.Timeout
	}
	return u
}

func (u *Unoserver) runServer(ctx context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.process != nil {
		return nil
	}

	slog.InfoContext(ctx, "Starting unoserver...")

	// nolint gosec
	cmd := exec.Command("unoserver", "--port", strconv.Itoa(u.port))

	if err := cmd.Start(); err != nil {
		return err
	}

	u.process = cmd.Process

	exitCh := make(chan error, 1)
	go func() {
		exitCh <- cmd.Wait()
		u.mu.Lock()
		u.process = nil
		u.mu.Unlock()
	}()

	timer := time.NewTimer(startWaitTime)
	defer timer.Stop()

	select {
	case err := <-exitCh:
		return err
	case <-timer.C:
		return nil
	}
}

// StopServer stops `unoserver`
func (u *Unoserver) StopServer(ctx context.Context) {
	u.mu.Lock()

	if u.process == nil {
		u.mu.Unlock()
		return
	}

	slog.InfoContext(ctx, "Shutting down unoserver...")

	proc := u.process
	u.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), shutdownWaitTime)
	defer cancel()

	process.ShutdownProcess(ctx, proc)

	slog.InfoContext(ctx, "Unoserver stopped")
}

var _ converter.Client = (*Unoserver)(nil)

// Convert converts source file to target file
func (u *Unoserver) Convert(ctx context.Context, from, to string, opts converter.ConvertOptions) error {
	u.semaphore <- struct{}{}
	defer func() {
		<-u.semaphore
	}()

	if u.process == nil {
		if err := u.runServer(ctx); err != nil {
			return fmt.Errorf("running unoserver: %w", err)
		}
	}

	args := []string{"--port", strconv.Itoa(u.port)}
	if opts.Filter != "" {
		args = append(args, "--filter", opts.Filter)
	}
	args = append(args, from, to)

	retryMinTimeout := time.Duration(0)

	return retry.Do(
		ctx,
		u.conversionRetries,
		retryMinTimeout,
		func() error {
			cmdCtx, cancel := context.WithTimeout(ctx, u.timeout)
			defer cancel()

			cmd := exec.CommandContext(cmdCtx, "unoconvert", args...)
			err := cmd.Start()
			if err != nil {
				return err
			}
			err = cmd.Wait()
			if err != nil {
				return err
			}
			return nil
		})
}
