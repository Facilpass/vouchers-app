package auth

import (
	"testing"
	"time"
)

func TestRateLimitAllow(t *testing.T) {
	rl := NewLoginRateLimiter(5, 15*time.Minute)
	for i := 0; i < 5; i++ {
		if !rl.Allow("1.2.3.4") {
			t.Fatalf("tentativa %d deveria passar", i+1)
		}
	}
}

func TestRateLimitBlock(t *testing.T) {
	rl := NewLoginRateLimiter(5, 15*time.Minute)
	for i := 0; i < 5; i++ {
		rl.Allow("1.2.3.4")
	}
	if rl.Allow("1.2.3.4") {
		t.Error("tentativa 6 deveria ser bloqueada")
	}
}

func TestRateLimitPerIP(t *testing.T) {
	rl := NewLoginRateLimiter(2, 15*time.Minute)
	rl.Allow("1.1.1.1")
	rl.Allow("1.1.1.1")
	if !rl.Allow("2.2.2.2") {
		t.Error("IP diferente deveria passar")
	}
}
