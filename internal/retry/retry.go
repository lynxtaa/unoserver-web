// Package retry provides helpers for retrying
package retry

import (
	"context"
	"fmt"
	"time"
)

// Do executes the effector function and retries on failure using exponential backoff.
func Do(
	ctx context.Context,
	attempts int,
	baseDelay time.Duration,
	fn func() error,
) error {
	var err error
	delay := baseDelay

	for i := range attempts {
		if err := ctx.Err(); err != nil {
			return err
		}

		err = fn()
		if err == nil {
			return nil
		}

		if i == attempts-1 {
			break
		}

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}

		delay *= 2
	}

	return fmt.Errorf("operation failed after %d attempts; last error: %w", attempts, err)
}
