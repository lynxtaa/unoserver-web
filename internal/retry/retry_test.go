package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func FailTimes(n int) func() error {
	attempt := 0

	return func() error {
		attempt++
		if attempt <= n {
			return errors.New("running func")
		}
		return nil
	}
}

func TestRetry(t *testing.T) {
	table := []struct {
		failTimes int
		retries   int
	}{
		{failTimes: 0, retries: 1},
		{failTimes: 1, retries: 1},
		{failTimes: 2, retries: 1},
	}

	minTimeout := 100 * time.Millisecond

	for _, st := range table {
		shouldError := st.failTimes > st.retries
		text := "no error"
		if shouldError {
			text = "error"
		}
		t.Run(
			fmt.Sprintf("%s if fn fails %d times and retrying %d times", text, st.failTimes, st.retries),
			func(t *testing.T) {
				err := Do(context.Background(), st.retries, minTimeout, FailTimes(st.failTimes))
				if shouldError && err == nil {
					t.Errorf("expected error, got no error")
				}
				if !shouldError && err != nil {
					t.Errorf("expected no error, got error")
				}
			})
	}

}
