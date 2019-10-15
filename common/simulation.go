package util

import (
	"math/rand"
	"time"
)

// Simulates a random delay between 0 and maxDelay milliseconds with a
// sleep.
func simulateDelay(maxDelay int) {
	time.Sleep(time.Duration(rand.Intn(maxDelay)) * time.Millisecond)
}
