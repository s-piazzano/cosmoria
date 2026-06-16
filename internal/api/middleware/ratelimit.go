package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	mu          sync.RWMutex
	requests    map[string][]time.Time
	maxRequests int
	window      time.Duration
}

func newRateLimiter(maxRequests int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests:    make(map[string][]time.Time),
		maxRequests: maxRequests,
		window:      window,
	}
}

func (rl *rateLimiter) Allow(ip string) bool {
	now := time.Now()
	cutoff := now.Add(-rl.window)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	timestamps := rl.requests[ip]
	var valid []time.Time
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	if len(valid) >= rl.maxRequests {
		rl.requests[ip] = valid
		return false
	}

	valid = append(valid, now)
	rl.requests[ip] = valid
	return true
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if host, _, err := net.SplitHostPort(xff); err == nil {
			return host
		}
		return xff
	}
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// RateLimit returns a middleware that limits requests per IP within a time window.
func RateLimit(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	rl := newRateLimiter(maxRequests, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getClientIP(r)
			if !rl.Allow(ip) {
				w.Header().Set("Retry-After", window.String())
				writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "rate_limit_exceeded"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
