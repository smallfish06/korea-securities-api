package kiwoom

import (
	"sync"
	"time"
)

// RateLimiter implements a token-bucket rate limiter.
type RateLimiter struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastTime   time.Time
}

// NewRateLimiter creates a new limiter with maxPerSecond throughput.
func NewRateLimiter(maxPerSecond float64) *RateLimiter {
	maxTok := maxPerSecond
	if maxTok < 1 {
		maxTok = 1
	}
	return &RateLimiter{
		tokens:     maxTok,
		maxTokens:  maxTok,
		refillRate: maxPerSecond,
		lastTime:   time.Now(),
	}
}

// Wait blocks until one token is available.
func (rl *RateLimiter) Wait() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastTime).Seconds()
	rl.tokens = minFloat(rl.maxTokens, rl.tokens+elapsed*rl.refillRate)
	rl.lastTime = now

	if rl.tokens < 1 {
		wait := time.Duration((1-rl.tokens)/rl.refillRate*float64(time.Second)) + time.Millisecond
		rl.mu.Unlock()
		time.Sleep(wait)
		rl.mu.Lock()
		rl.tokens = 0
		rl.lastTime = time.Now()
		return
	}

	rl.tokens--
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
