package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"hostaggr/internal/search"
)

func TestRateLimiter_AllowRequestsWithinLimit(t *testing.T) {
	rl := search.NewRateLimiter(10, 1*time.Minute)
	ip := "192.168.1.1"

	// The rate limiter allows 10 requests per minute
	// All 10 requests should be allowed
	for i := 0; i < 10; i++ {
		allowed := rl.Allow(ip)
		assert.True(t, allowed, "request %d should be allowed (within limit of 10)", i+1)
	}
}

func TestRateLimiter_BlockRequestsOverLimit(t *testing.T) {
	rl := search.NewRateLimiter(10, 1*time.Minute)
	ip := "192.168.1.2"

	// Use up all 10 tokens
	for i := 0; i < 10; i++ {
		allowed := rl.Allow(ip)
		assert.True(t, allowed, "request %d should be allowed", i+1)
	}

	// The 11th request should be blocked
	allowed := rl.Allow(ip)
	assert.False(t, allowed, "request 11 should be blocked (over limit)")

	// Additional requests should also be blocked
	for i := 0; i < 5; i++ {
		allowed := rl.Allow(ip)
		assert.False(t, allowed, "additional request %d should be blocked", i+1)
	}
}

func TestRateLimiter_TokenRefillOverTime(t *testing.T) {
	rl := search.NewRateLimiter(10, 1*time.Minute)
	ip := "192.168.1.3"

	// Use up all 10 tokens
	for i := 0; i < 10; i++ {
		allowed := rl.Allow(ip)
		assert.True(t, allowed, "initial request %d should be allowed", i+1)
	}

	// Next request should be blocked
	allowed := rl.Allow(ip)
	assert.False(t, allowed, "should be blocked immediately after using all tokens")

	// Wait for some time to allow token refill
	// The refill rate is 10 tokens per minute (1 minute = 60 seconds)
	// So we should get approximately 1 token every 6 seconds
	// Let's wait 7 seconds to be safe and ensure at least 1 token is refilled
	time.Sleep(7 * time.Second)

	// After waiting, we should have at least 1 token refilled
	allowed = rl.Allow(ip)
	assert.True(t, allowed, "should be allowed after token refill")

	// But the next one should be blocked again (only 1 token was refilled)
	allowed = rl.Allow(ip)
	assert.False(t, allowed, "should be blocked again after using refilled token")
}

func TestRateLimiter_PartialRefill(t *testing.T) {
	rl := search.NewRateLimiter(10, 1*time.Minute)
	ip := "192.168.1.4"

	// Use up all 10 tokens
	for i := 0; i < 10; i++ {
		allowed := rl.Allow(ip)
		assert.True(t, allowed)
	}

	// Wait for half the refill period (30 seconds = half of 1 minute)
	// This should refill approximately 5 tokens
	time.Sleep(30 * time.Second)

	// We should be able to make approximately 5 more requests
	successCount := 0
	for i := 0; i < 7; i++ {
		if rl.Allow(ip) {
			successCount++
		}
	}

	// We expect around 5 successful requests (±1 for timing variations)
	assert.GreaterOrEqual(t, successCount, 4, "should allow at least 4 requests after 30s")
	assert.LessOrEqual(t, successCount, 6, "should allow at most 6 requests after 30s")
}

func TestRateLimiter_MultipleIPsTrackedSeparately(t *testing.T) {
	rl := search.NewRateLimiter(10, 1*time.Minute)

	ip1 := "192.168.1.10"
	ip2 := "192.168.1.11"
	ip3 := "10.0.0.1"

	// Each IP should have its own bucket with 10 tokens

	// IP1: use 10 tokens
	for i := 0; i < 10; i++ {
		allowed := rl.Allow(ip1)
		assert.True(t, allowed, "IP1 request %d should be allowed", i+1)
	}

	// IP1: 11th request should be blocked
	allowed := rl.Allow(ip1)
	assert.False(t, allowed, "IP1 should be blocked after 10 requests")

	// IP2: should still have all 10 tokens available
	for i := 0; i < 10; i++ {
		allowed := rl.Allow(ip2)
		assert.True(t, allowed, "IP2 request %d should be allowed", i+1)
	}

	// IP2: 11th request should be blocked
	allowed = rl.Allow(ip2)
	assert.False(t, allowed, "IP2 should be blocked after 10 requests")

	// IP3: should still have all 10 tokens available
	for i := 0; i < 10; i++ {
		allowed := rl.Allow(ip3)
		assert.True(t, allowed, "IP3 request %d should be allowed", i+1)
	}

	// IP3: 11th request should be blocked
	allowed = rl.Allow(ip3)
	assert.False(t, allowed, "IP3 should be blocked after 10 requests")

	// Verify IP1 is still blocked
	allowed = rl.Allow(ip1)
	assert.False(t, allowed, "IP1 should still be blocked")
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	// Run with: go test -race
	rl := search.NewRateLimiter(10, 1*time.Minute)

	numGoroutines := 50
	numRequests := 20

	var wg sync.WaitGroup
	var mu sync.Mutex
	allowedCount := 0

	// Multiple goroutines making requests from the same IP
	ip := "192.168.1.20"

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numRequests; j++ {
				if rl.Allow(ip) {
					mu.Lock()
					allowedCount++
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	// Only 10 requests should have been allowed total (the initial bucket size)
	assert.Equal(t, 10, allowedCount, "only 10 requests should be allowed across all goroutines")
}

func TestRateLimiter_ConcurrentMultipleIPs(t *testing.T) {
	// Run with: go test -race
	rl := search.NewRateLimiter(10, 1*time.Minute)

	numIPs := 20
	numRequestsPerIP := 15

	var wg sync.WaitGroup
	results := make(map[string]int)
	var mu sync.Mutex

	for i := 0; i < numIPs; i++ {
		wg.Add(1)
		ip := "192.168.1." + string(rune(100+i))

		go func(ipAddr string) {
			defer wg.Done()
			allowed := 0
			for j := 0; j < numRequestsPerIP; j++ {
				if rl.Allow(ipAddr) {
					allowed++
				}
			}

			mu.Lock()
			results[ipAddr] = allowed
			mu.Unlock()
		}(ip)
	}

	wg.Wait()

	// Each IP should have had exactly 10 requests allowed
	for ip, count := range results {
		assert.Equal(t, 10, count, "IP %s should have exactly 10 allowed requests", ip)
	}
}

func TestRateLimiter_RefillDoesNotExceedMax(t *testing.T) {
	rl := search.NewRateLimiter(10, 1*time.Minute)
	ip := "192.168.1.30"

	// Don't use any tokens initially
	// Wait for longer than the refill period
	time.Sleep(65 * time.Second)

	// Should still only have 10 tokens (not more)
	allowedCount := 0
	for i := 0; i < 15; i++ {
		if rl.Allow(ip) {
			allowedCount++
		}
	}

	assert.Equal(t, 10, allowedCount, "should not have more than 10 tokens even after waiting")
}

func TestRateLimiter_NewIPStartsWithFullBucket(t *testing.T) {
	rl := search.NewRateLimiter(10, 1*time.Minute)

	// Test multiple new IPs
	ips := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3", "192.168.1.1"}

	for _, ip := range ips {
		// Each new IP should start with 10 tokens
		allowedCount := 0
		for i := 0; i < 15; i++ {
			if rl.Allow(ip) {
				allowedCount++
			}
		}

		assert.Equal(t, 10, allowedCount, "new IP %s should start with exactly 10 tokens", ip)
	}
}

func TestRateLimiter_BurstTraffic(t *testing.T) {
	rl := search.NewRateLimiter(10, 1*time.Minute)
	ip := "192.168.1.40"

	// Simulate burst traffic - make 20 requests as fast as possible
	allowedCount := 0
	blockedCount := 0

	for i := 0; i < 20; i++ {
		if rl.Allow(ip) {
			allowedCount++
		} else {
			blockedCount++
		}
	}

	assert.Equal(t, 10, allowedCount, "should allow exactly 10 requests in burst")
	assert.Equal(t, 10, blockedCount, "should block exactly 10 requests in burst")
}

func TestRateLimiter_GradualRefill(t *testing.T) {
	rl := search.NewRateLimiter(10, 1*time.Minute)
	ip := "192.168.1.50"

	// Use all tokens
	for i := 0; i < 10; i++ {
		rl.Allow(ip)
	}

	// Should be blocked now
	assert.False(t, rl.Allow(ip), "should be blocked after using all tokens")

	// Wait 6 seconds (should refill ~1 token)
	time.Sleep(6 * time.Second)
	assert.True(t, rl.Allow(ip), "should allow 1 request after 6 seconds")
	assert.False(t, rl.Allow(ip), "should block next request")

	// Wait another 6 seconds
	time.Sleep(6 * time.Second)
	assert.True(t, rl.Allow(ip), "should allow 1 request after another 6 seconds")
	assert.False(t, rl.Allow(ip), "should block next request")
}

func TestRateLimiter_FullRecovery(t *testing.T) {
	rl := search.NewRateLimiter(10, 1*time.Minute)
	ip := "192.168.1.60"

	// Use all tokens
	for i := 0; i < 10; i++ {
		rl.Allow(ip)
	}

	// Verify blocked
	assert.False(t, rl.Allow(ip), "should be blocked")

	// Wait for full refill (60 seconds)
	time.Sleep(61 * time.Second)

	// Should have full bucket again
	allowedCount := 0
	for i := 0; i < 15; i++ {
		if rl.Allow(ip) {
			allowedCount++
		}
	}

	assert.Equal(t, 10, allowedCount, "should have full bucket (10 tokens) after 60 seconds")
}
