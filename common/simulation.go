package util

import (
	"time"
)

// Simulates a delay of given milliseconds with a
// sleep.
func SimulateDelay(delay int) {
	time.Sleep(time.Duration(delay) * time.Millisecond)
}
