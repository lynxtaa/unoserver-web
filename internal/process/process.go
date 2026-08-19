// Package process provides process management utilities
package process

import (
	"context"
	"os"
)

// ShutdownProcess gracefully shuts down a process by sending an interrupt signal
func ShutdownProcess(ctx context.Context, proc *os.Process) {
	if err := proc.Signal(os.Interrupt); err != nil {
		return
	}

	<-ctx.Done()

	if ctx.Err() == context.DeadlineExceeded {
		_ = proc.Kill()
	}
}
