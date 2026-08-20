// Package unoserver provides Python's `unoserver`
// for handling LibreOffice conversions
package unoserver

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lynxtaa/unoserver-web/internal/converter"
)

const (
	startWaitTime     = 30 * time.Second
	shutdownWaitTime  = 10 * time.Second
	retryBaseDelay    = 1 * time.Second
	defaultMaxWorkers = 8
)

// Unoserver contains everything related to `unoserver`
type Unoserver struct {
	semaphore         chan struct{}
	timeout           time.Duration
	conversionRetries int
	port              int
	mu                sync.Mutex
	process           *os.Process
	// done is closed once the running process has exited
	done chan struct{}
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
	maxWorkers := opts.MaxWorkers
	if maxWorkers <= 0 {
		maxWorkers = defaultMaxWorkers
	}

	u := &Unoserver{
		semaphore:         make(chan struct{}, maxWorkers),
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

	// New context is intentional otherwise unoserver will be killed after each request
	cmd := exec.CommandContext(context.Background(), "unoserver", "--port", strconv.Itoa(u.port))

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan struct{})
	u.process = cmd.Process
	u.done = done

	errCh := make(chan error, 1)

	go func() {
		errCh <- cmd.Wait()
		close(done)

		u.mu.Lock()
		defer u.mu.Unlock()
		if u.process == cmd.Process {
			u.process = nil
			u.done = nil
		}
	}()

	startedCh := make(chan struct{})

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), "unoserver:Started") {
				close(startedCh)
				break
			}
		}
		_ = scanner.Err()
		// Consume the rest of stdout to prevent the process from blocking
		// when the OS pipe buffer fills up.
		_, _ = io.Copy(io.Discard, stderr)
	}()

	select {
	case <-startedCh:
		slog.InfoContext(ctx, "Unoserver started")
		return nil
	case err := <-errCh:
		return fmt.Errorf("unoserver exited prematurely: %w", err)
	case <-time.After(startWaitTime):
		_ = cmd.Process.Kill()
		return fmt.Errorf("timeout waiting for unoserver to start")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// StopServer stops `unoserver`
func (u *Unoserver) StopServer(ctx context.Context) {
	u.mu.Lock()
	proc, done := u.process, u.done
	u.mu.Unlock()

	if proc == nil {
		return
	}

	slog.InfoContext(ctx, "Shutting down unoserver...")

	// ctx is usually already cancelled by a signal, the process still gets its grace period
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownWaitTime)
	defer cancel()

	if err := proc.Signal(os.Interrupt); err != nil {
		// Already gone
		return
	}

	select {
	case <-done:
	case <-shutdownCtx.Done():
		slog.WarnContext(ctx, "Unoserver didn't exit in time, killing it")
		_ = proc.Kill()
	}

	slog.InfoContext(ctx, "Unoserver stopped")
}

// Convert converts source file to target file
func (u *Unoserver) Convert(ctx context.Context, from, to string, opts converter.ConvertOptions) error {
	// Queued requests give up their slot as soon as the client is gone
	select {
	case u.semaphore <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}

	defer func() {
		<-u.semaphore
	}()

	args := []string{"--port", strconv.Itoa(u.port)}
	if opts.Filter != "" {
		args = append(args, "--filter", opts.Filter)
	}
	args = append(args, from, to)

	var lastErr error

	for attempt := range u.conversionRetries + 1 {
		if err := ctx.Err(); err != nil {
			return err
		}

		if attempt > 0 {
			// Exponential backoff, unoserver may still be recovering
			select {
			case <-time.After(retryBaseDelay << (attempt - 1)):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Checked on every attempt: unoserver could have died in between
		if err := u.runServer(ctx); err != nil {
			lastErr = fmt.Errorf("running unoserver: %w", err)
			continue
		}

		lastErr = u.unoconvert(ctx, args)
		if lastErr == nil {
			return nil
		}
	}

	return fmt.Errorf("operation failed after %d retries; last error: %w", u.conversionRetries, lastErr)
}

func (u *Unoserver) unoconvert(ctx context.Context, args []string) error {
	cmdCtx, cancel := context.WithTimeout(ctx, u.timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "unoconvert", args...)
	err := cmd.Run()
	if err != nil {
		if e := (&exec.ExitError{}); errors.As(err, &e) {
			slog.ErrorContext(ctx, "unoconvert exited with non-zero code", "stderr", string(e.Stderr))
		}
		return err
	}
	return nil
}
