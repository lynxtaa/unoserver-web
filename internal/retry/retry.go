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
	retries int,
	baseDelay time.Duration,
	fn func() error,
) error {
	var err error
	delay := baseDelay

	for i := range retries + 1 {
		if err := ctx.Err(); err != nil {
			return err
		}

		err = fn()
		if err == nil {
			return nil
		}

		if i == retries {
			break
		}

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}

		delay *= 2
	}

	return fmt.Errorf("operation failed after %d retries; last error: %w", retries, err)
}
