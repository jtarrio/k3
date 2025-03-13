package k3

import "time"

// Clock is an interface for objects that return the current time.
type Clock interface {
	// Now returns the current time
	Now() time.Time
}

// NewSystemClock returns a Clock that uses the computer's clock.
func SystemClock() Clock {
	return &systemClock{}
}

type systemClock struct{}

func (c systemClock) Now() time.Time {
	return time.Now()
}

// NewIncreasingClock modifies the given clock so it always increases by at least 1 millisecond.
func NewIncreasingClock(clock Clock) Clock {
	return &increasingClock{parent: clock}
}

type increasingClock struct {
	parent Clock
	next   time.Time
}

func (c *increasingClock) Now() time.Time {
	now := c.parent.Now()
	if now.Before(c.next) {
		now = c.next
	}
	c.next = now.Add(1 * time.Millisecond)
	return now
}
