// Package ratelimit provides a shared rate limiter for broker API clients.
//
// It wraps golang.org/x/time/rate to provide context-aware, per-broker
// request throttling. Multiple callers sharing the same Limiter are
// serialized correctly without exceeding the configured rate.
package ratelimit

import (
	"context"
	"fmt"

	"golang.org/x/time/rate"
)

// Limiter throttles API requests using a token-bucket algorithm.
type Limiter struct {
	limiter *rate.Limiter
	name    string
}

// New creates a Limiter that allows rps requests per second with the given
// burst size. Burst controls how many requests can fire at once before
// throttling kicks in.
func New(name string, rps float64, burst int) *Limiter {
	if burst < 1 {
		burst = 1
	}
	return &Limiter{
		limiter: rate.NewLimiter(rate.Limit(rps), burst),
		name:    name,
	}
}

// Wait blocks until the limiter allows one request, or ctx is cancelled.
func (l *Limiter) Wait(ctx context.Context) error {
	if err := l.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limit [%s]: %w", l.name, err)
	}
	return nil
}

// Name returns the limiter's name (for logging/debugging).
func (l *Limiter) Name() string {
	return l.name
}

// Allow reports whether a request can proceed right now without waiting.
func (l *Limiter) Allow() bool {
	return l.limiter.Allow()
}
