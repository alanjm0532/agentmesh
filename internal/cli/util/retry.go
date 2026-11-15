package util

import (
	"context"
	"fmt"
	"time"
)

// Retry runs action at most maxAttempts times, including the initial attempt.
// shouldRetry selects recoverable errors. Retries wait with exponential backoff,
// starting at 200 milliseconds and capped at 5 seconds, respecting cancellation.
func Retry(ctx context.Context, maxAttempts int, shouldRetry func(error) bool, action func() error) error {
	if maxAttempts < 1 {
		return fmt.Errorf("retry attempts must be at least 1")
	}
	delay := 200 * time.Millisecond
	for attempt := 0; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := action()
		if err == nil || attempt == maxAttempts-1 || !shouldRetry(err) {
			return err
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		delay = min(delay*2, 5*time.Second)
	}
}
