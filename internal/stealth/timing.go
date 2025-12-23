package stealth

import (
	"math/rand"
	"time"
)

// RandomDelay sleeps for a random duration between min and max. This is used
// to introduce "think time" between actions to better approximate human
// behaviour.
func RandomDelay(min, max time.Duration) {
	if max <= 0 || max < min {
		time.Sleep(min)
		return
	}
	delta := max - min
	sleep := min + time.Duration(rand.Int63n(int64(delta)))
	time.Sleep(sleep)
}



