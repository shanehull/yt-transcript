package yt_transcript

import (
	"sync"
	"time"
)

// breaker trips after consecutive upstream throttles so a blocked client stops
// hammering YouTube. Once open it admits a single probe after a cooldown that
// doubles on each failed probe. Success resets it.
type breaker struct {
	mu           sync.Mutex
	threshold    int
	baseCooldown time.Duration
	maxCooldown  time.Duration

	failures  int
	opens     int
	openUntil time.Time
}

func newBreaker(threshold int, baseCooldown, maxCooldown time.Duration) *breaker {
	if threshold < 1 {
		threshold = 1
	}
	return &breaker{threshold: threshold, baseCooldown: baseCooldown, maxCooldown: maxCooldown}
}

// allow reports whether a request may proceed. When blocked it returns the time
// remaining until the next probe is permitted.
func (b *breaker) allow(now time.Time) (time.Duration, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.openUntil.IsZero() && now.Before(b.openUntil) {
		return b.openUntil.Sub(now), false
	}
	return 0, true
}

func (b *breaker) succeed() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.opens = 0
	b.openUntil = time.Time{}
}

// fail records a throttle. retryAfter, when positive, is used verbatim as the
// cooldown; otherwise the cooldown doubles on each successive open. It returns
// the cooldown applied, or zero while still below the failure threshold.
func (b *breaker) fail(retryAfter time.Duration, now time.Time) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.opens == 0 {
		b.failures++
		if b.failures < b.threshold {
			return 0
		}
	}

	cooldown := retryAfter
	if cooldown <= 0 {
		cooldown = b.baseCooldown << b.opens
		if cooldown <= 0 || cooldown > b.maxCooldown {
			cooldown = b.maxCooldown
		}
	}
	b.openUntil = now.Add(cooldown)
	b.opens++
	b.failures = 0
	return cooldown
}
