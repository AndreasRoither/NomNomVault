package clock

import "time"

// Clock returns the current time.
type Clock interface {
	Now() time.Time
}

// RealClock uses time.Now.
type RealClock struct{}

// Now returns the current wall time.
func (RealClock) Now() time.Time {
	return time.Now().UTC()
}
