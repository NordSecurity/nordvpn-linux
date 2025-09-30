package tray

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

type BackoffConfig struct {
	// InitialDelay time for the first retry.
	// Will default to 1 second if non-positive value is provided.
	InitialDelay time.Duration

	// MaxDelay maximum time between attempts.
	// Will default to 1 minute if non-positive value is provided.
	MaxDelay time.Duration

	// MaxRetries decides number of times operation will be attempted.
	// <=0 means unlimited retries until context is cancelled.
	MaxRetries int
}

// DefaultBackoffConfig provides default backoff configuration
func DefaultBackoffConfig() BackoffConfig {
	return BackoffConfig{
		InitialDelay: time.Second,
		MaxDelay:     time.Minute,
		MaxRetries:   7,
	}
}

// RetryWithBackoff attempts to execute an operation, retrying with exponential backoff on failure.
func RetryWithBackoff(
	ctx context.Context,
	cfg BackoffConfig,
	op func(ctx context.Context) error,
) error {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	retryDelay := cfg.InitialDelay
	if retryDelay <= 0 {
		retryDelay = 1 * time.Second
	}

	maxDelay := cfg.MaxDelay
	if maxDelay <= 0 {
		maxDelay = time.Minute
	}

	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 0
	}

	var lastErr error
	for i := 0; maxRetries == 0 || i < maxRetries; i++ {
		err := op(ctx)
		if err == nil {
			return nil
		}
		lastErr = err

		if ctx.Err() != nil {
			return fmt.Errorf("context cancelled during operation: %w", ctx.Err())
		}

		// Exit if this was the last attempt
		if maxRetries > 0 && i == maxRetries-1 {
			break
		}

		// Jitter ±10%
		jitter := 0.9 + r.Float64()*0.2
		jitteredDelay := time.Duration(float64(retryDelay) * jitter)

		delayTimer := time.NewTimer(jitteredDelay)
		select {
		case <-ctx.Done():
			delayTimer.Stop()
			return fmt.Errorf("context cancelled during backoff: %w", ctx.Err())
		case <-delayTimer.C:
		}

		retryDelay *= 2
		if retryDelay > maxDelay {
			retryDelay = maxDelay
		}
	}

	return fmt.Errorf("operation failed after %d attempts: %w", maxRetries, lastErr)
}
