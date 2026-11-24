package middleware

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	// Create a limiter: 10 requests per second, max burst of 10
	rl := NewRateLimiter(10, time.Second, 10)

	// First 10 requests should be allowed (burst)
	for i := 0; i < 10; i++ {
		if !rl.Allow("test-key") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 11th request should be denied
	if rl.Allow("test-key") {
		t.Error("Request 11 should be denied")
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	// Create a limiter: 1 request per second, max burst of 2
	rl := NewRateLimiter(1, time.Second, 2)

	// Use up burst
	if !rl.Allow("test-key") {
		t.Error("First request should be allowed")
	}
	if !rl.Allow("test-key") {
		t.Error("Second request should be allowed (burst)")
	}

	// Should be denied now
	if rl.Allow("test-key") {
		t.Error("Third request should be denied")
	}

	// Wait for refill
	time.Sleep(1100 * time.Millisecond)

	// Should be allowed again
	if !rl.Allow("test-key") {
		t.Error("Request after refill should be allowed")
	}
}

func TestRateLimiter_MultipleKeys(t *testing.T) {
	rl := NewRateLimiter(1, time.Second, 2)

	// Use key1's burst
	rl.Allow("key1")
	rl.Allow("key1")

	// key2 should still have its own burst
	if !rl.Allow("key2") {
		t.Error("key2 should have its own burst")
	}
}

func TestRateLimiter_BurstRecovery(t *testing.T) {
	// Create a limiter: 1 request per second, max burst of 5
	rl := NewRateLimiter(1, 100*time.Millisecond, 5)

	// Use burst
	for i := 0; i < 5; i++ {
		rl.Allow("test")
	}

	// Wait for some refill
	time.Sleep(250 * time.Millisecond)

	// Should have gotten ~2-3 tokens back
	allowed := 0
	for i := 0; i < 3; i++ {
		if rl.Allow("test") {
			allowed++
		}
	}

	if allowed < 2 {
		t.Errorf("Should have gotten at least 2 tokens back, got %d", allowed)
	}
}

func TestNewRateLimiter(t *testing.T) {
	rl := NewRateLimiter(100, time.Minute, 100)

	if rl.rate != 100 {
		t.Errorf("rate = %d, want 100", rl.rate)
	}
	if rl.interval != time.Minute {
		t.Errorf("interval = %v, want 1m", rl.interval)
	}
	if rl.maxTokens != 100 {
		t.Errorf("maxTokens = %d, want 100", rl.maxTokens)
	}
}
