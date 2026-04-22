package auth

import (
	"sync"
	"time"
)

type LoginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	max      int
	window   time.Duration
}

func NewLoginRateLimiter(max int, window time.Duration) *LoginRateLimiter {
	return &LoginRateLimiter{
		attempts: make(map[string][]time.Time),
		max:      max,
		window:   window,
	}
}

func (rl *LoginRateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)
	valid := make([]time.Time, 0, len(rl.attempts[ip]))
	for _, t := range rl.attempts[ip] {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	if len(valid) >= rl.max {
		rl.attempts[ip] = valid
		return false
	}
	valid = append(valid, now)
	rl.attempts[ip] = valid
	return true
}
