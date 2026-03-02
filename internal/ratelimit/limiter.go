package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	rate     rate.Limit
	burst    int
}

func NewIPRateLimiter(requestsPerMinute, burst int) *IPRateLimiter {
	r := rate.Limit(float64(requestsPerMinute) / 60.0) // Convert to per-second rate
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     r,
		burst:    burst,
	}
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(i.rate, i.burst)
		i.limiters[ip] = limiter
	}

	return limiter
}

// Cleanup removes old limiters periodically
func (i *IPRateLimiter) Cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			i.mu.Lock()
			// Simple cleanup: clear all limiters periodically
			// In production, you might want more sophisticated cleanup
			i.limiters = make(map[string]*rate.Limiter)
			i.mu.Unlock()
		}
	}()
}
