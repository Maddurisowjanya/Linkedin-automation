package stealth

import (
	"math"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// ScrollHumanLike performs a series of scroll operations with random
// distances, natural acceleration / deceleration and occasional scroll‑back.
// Combined with RandomDelay this gives pages time to load and mimics how
// humans skim content.
func ScrollHumanLike(page *rod.Page, totalDuration time.Duration) error {
	start := time.Now()
	for time.Since(start) < totalDuration {
		// Random direction and magnitude.
		dir := 1.0
		if rand.Float64() < 0.2 {
			dir = -1.0 // occasional scroll‑back
		}
		maxDelta := 400 + rand.Intn(400)

		steps := 10 + rand.Intn(15)
		for i := 0; i < steps; i++ {
			progress := float64(i) / float64(steps)
			// Easing function to accelerate then decelerate.
			ease := math.Sin(progress * math.Pi)
			delta := int(float64(maxDelta) * ease * dir / float64(steps))
			if delta == 0 {
				continue
			}
			// Use a wheel event via DevTools protocol to perform scrolling.
			_ = proto.InputDispatchMouseEvent{
				Type:   proto.InputDispatchMouseEventTypeMouseWheel,
				DeltaX: 0,
				DeltaY: float64(delta),
			}.Call(page)
			time.Sleep(time.Duration(15+rand.Intn(30)) * time.Millisecond)
		}

		// Random pause between bursts of scrolling.
		RandomDelay(500*time.Millisecond, 2*time.Second)
	}
	return nil
}



