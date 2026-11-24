package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter implements a simple token bucket rate limiter
type RateLimiter struct {
	mu          sync.RWMutex
	buckets     map[string]*bucket
	rate        int           // tokens per interval
	interval    time.Duration // refill interval
	maxTokens   int           // maximum tokens
	cleanupTick time.Duration
}

type bucket struct {
	tokens     int
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
// rate: number of requests allowed per interval
// interval: time period for rate calculation
// maxTokens: maximum burst size
func NewRateLimiter(rate int, interval time.Duration, maxTokens int) *RateLimiter {
	rl := &RateLimiter{
		buckets:     make(map[string]*bucket),
		rate:        rate,
		interval:    interval,
		maxTokens:   maxTokens,
		cleanupTick: interval * 10,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request is allowed for the given key
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.buckets[key]
	now := time.Now()

	if !exists {
		rl.buckets[key] = &bucket{
			tokens:     rl.maxTokens - 1, // consume one token
			lastRefill: now,
		}
		return true
	}

	// Refill tokens based on time passed
	elapsed := now.Sub(b.lastRefill)
	refills := int(elapsed / rl.interval)

	if refills > 0 {
		b.tokens += refills * rl.rate
		if b.tokens > rl.maxTokens {
			b.tokens = rl.maxTokens
		}
		b.lastRefill = now
	}

	// Check if we have tokens
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// cleanup removes stale buckets periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupTick)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, b := range rl.buckets {
			// Remove buckets that haven't been used for 10x the interval
			if now.Sub(b.lastRefill) > rl.interval*10 {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit returns a Gin middleware that rate limits by IP
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use IP as the key
		key := c.ClientIP()

		if !rl.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": rl.interval.String(),
			})
			return
		}

		c.Next()
	}
}

// IPRateLimiter creates a rate limiter that uses client IP
func IPRateLimiter(requestsPerMinute int) gin.HandlerFunc {
	limiter := NewRateLimiter(
		requestsPerMinute/60,   // tokens per second
		time.Second,            // interval
		requestsPerMinute/30+1, // burst size
	)
	return limiter.RateLimit()
}

// APIKeyRateLimiter creates a rate limiter that uses API key
func APIKeyRateLimiter(requestsPerMinute int) gin.HandlerFunc {
	limiter := NewRateLimiter(
		requestsPerMinute/60,   // tokens per second
		time.Second,            // interval
		requestsPerMinute/30+1, // burst size
	)
	return func(c *gin.Context) {
		// Use API key or IP as fallback
		key := c.GetHeader("Authorization")
		if key == "" {
			key = c.ClientIP()
		}

		if !limiter.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": time.Minute.String(),
			})
			return
		}

		c.Next()
	}
}

// EndpointRateLimiter creates per-endpoint rate limiters
type EndpointRateLimiter struct {
	limiters       map[string]*RateLimiter
	defaultLimiter *RateLimiter
	mu             sync.RWMutex
}

// NewEndpointRateLimiter creates a new endpoint-based rate limiter
func NewEndpointRateLimiter(defaultRate int) *EndpointRateLimiter {
	return &EndpointRateLimiter{
		limiters: make(map[string]*RateLimiter),
		defaultLimiter: NewRateLimiter(
			defaultRate/60,
			time.Second,
			defaultRate/30+1,
		),
	}
}

// AddEndpoint adds a rate limiter for a specific endpoint
func (erl *EndpointRateLimiter) AddEndpoint(endpoint string, rate int) {
	erl.mu.Lock()
	defer erl.mu.Unlock()

	erl.limiters[endpoint] = NewRateLimiter(
		rate/60,
		time.Second,
		rate/30+1,
	)
}

// Middleware returns a Gin middleware for endpoint-based rate limiting
func (erl *EndpointRateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		erl.mu.RLock()
		limiter, exists := erl.limiters[path]
		if !exists {
			limiter = erl.defaultLimiter
		}
		erl.mu.RUnlock()

		key := c.ClientIP()
		if apiKey := c.GetHeader("Authorization"); apiKey != "" {
			key = apiKey
		}

		if !limiter.Allow(key) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": time.Minute.String(),
			})
			return
		}

		c.Next()
	}
}
