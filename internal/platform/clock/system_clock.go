package clock

import "time"

type SystemClock struct{}

func NewSystemClock() *SystemClock { return &SystemClock{} }

func (c *SystemClock) Now() time.Time { return time.Now() }
