package ratelimit

import (
	"sync"
	"time"
)

type entry struct {
	count   int
	resetAt time.Time
}

// Limiter applies a fixed-window request cap per key.
type Limiter struct {
	mu      sync.Mutex
	max     int
	window  time.Duration
	entries map[string]entry
}

// New creates a limiter for the supplied fixed window.
func New(max int, window time.Duration) *Limiter {
	return &Limiter{
		max:     max,
		window:  window,
		entries: make(map[string]entry),
	}
}

// Allow reports whether a request should proceed and the remaining retry delay when denied.
func (l *Limiter) Allow(key string, now time.Time) (bool, time.Duration) {
	if l == nil || l.max <= 0 || l.window <= 0 {
		return true, 0
	}
	if key == "" {
		key = "global"
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for candidate, current := range l.entries {
		if now.After(current.resetAt) {
			delete(l.entries, candidate)
		}
	}

	current := l.entries[key]
	if current.resetAt.IsZero() || now.After(current.resetAt) {
		current = entry{
			count:   1,
			resetAt: now.Add(l.window),
		}
		l.entries[key] = current
		return true, 0
	}

	if current.count >= l.max {
		return false, time.Until(current.resetAt)
	}

	current.count++
	l.entries[key] = current
	return true, 0
}
