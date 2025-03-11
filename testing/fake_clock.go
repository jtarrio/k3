package testing

import "time"

// FakeClock is an atp.Clock that returns a user-provided time.
type FakeClock struct {
	Time time.Time
}

func (f *FakeClock) Now() time.Time {
	return f.Time
}
