package search

import (
	"sync"
	"time"
)

// bucket represents a token bucket for a single IP address
type bucket struct {
	tokens     int
	lastRefill time.Time
}

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu         sync.Mutex
	buckets    map[string]*bucket
	maxTokens  int
	refillRate time.Duration
}

// NewRateLimiter creates a new rate limiter that allows 10 requests per minute per IP
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		buckets:    make(map[string]*bucket),
		maxTokens:  10,
		refillRate: 1 * time.Minute,
	}

	// Start background cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request from the given IP address should be allowed
// Returns true if the request is allowed, false if rate limit exceeded
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Get or create bucket for this IP
	b, exists := rl.buckets[ip]
	if !exists {
		b = &bucket{
			tokens:     rl.maxTokens,
			lastRefill: time.Now(),
		}
		rl.buckets[ip] = b
	}

	// Refill tokens based on time elapsed
	rl.refillBucket(b)

	// Check if we have tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// refillBucket calculates and adds tokens based on time elapsed since last refill
func (rl *RateLimiter) refillBucket(b *bucket) {
	now := time.Now()
	elapsed := now.Sub(b.lastRefill)

	// Calculate how many tokens to add based on elapsed time
	// refillRate is the time to fully refill the bucket
	tokensToAdd := int(elapsed.Seconds() / rl.refillRate.Seconds() * float64(rl.maxTokens))

	if tokensToAdd > 0 {
		b.tokens += tokensToAdd
		if b.tokens > rl.maxTokens {
			b.tokens = rl.maxTokens
		}
		b.lastRefill = now
	}
}

// cleanup removes old buckets every 5 minutes to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()

		// Remove buckets that haven't been used in the last 10 minutes
		cutoff := time.Now().Add(-10 * time.Minute)
		for ip, b := range rl.buckets {
			if b.lastRefill.Before(cutoff) {
				delete(rl.buckets, ip)
			}
		}

		rl.mu.Unlock()
	}
}
