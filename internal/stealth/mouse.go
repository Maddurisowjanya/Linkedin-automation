package stealth

import (
	"math"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// MoveMouseHumanly moves the mouse cursor along a smooth, non‑linear Bézier
// curve with variable speed, micro‑corrections and a small overshoot before
// settling on the target. The goal is to roughly approximate how a human
// hand moves a mouse instead of teleporting in a straight line.
func MoveMouseHumanly(page *rod.Page, targetX, targetY float64) error {
	// For simplicity we assume a starting position close to the origin of
	// the viewport. The important part for this PoC is the *shape* and
	// timing of the movement, not the absolute starting coordinates.
	startX, startY := 10.0+rand.Float64()*20, 10.0+rand.Float64()*20

	// Introduce a slight overshoot past the target and then plan a tiny
	// corrective motion back to the final location.
	overshootX := targetX + (rand.Float64()-0.5)*20
	overshootY := targetY + (rand.Float64()-0.5)*20

	path := bezierPath(startX, startY, overshootX, overshootY, 80)
	for i, p := range path {
		// Variable speed: accelerate at the start, decelerate near the end.
		progress := float64(i) / float64(len(path))
		speedFactor := 0.5 + 1.5*math.Sin(progress*math.Pi) // bell‑curve speed
		// Occasional micro‑corrections to break perfect smoothness.
		p[0] += (rand.Float64() - 0.5) * 1.5
		p[1] += (rand.Float64() - 0.5) * 1.5

		// Dispatch a low‑level mouse move event via the Chrome DevTools
		// protocol. This does not rely on Rod's higher‑level mouse helpers,
		// but still results in realistic pointer movement from the page's
		// perspective.
		_ = proto.InputDispatchMouseEvent{
			Type: proto.InputDispatchMouseEventTypeMouseMoved,
			X:    p[0],
			Y:    p[1],
		}.Call(page)
		time.Sleep(time.Duration(4+rand.Intn(6)) * time.Millisecond / time.Duration(speedFactor))
	}

	// Small corrective move back to the precise target.
	_ = proto.InputDispatchMouseEvent{
		Type: proto.InputDispatchMouseEventTypeMouseMoved,
		X:    targetX,
		Y:    targetY,
	}.Call(page)
	time.Sleep(time.Duration(20+rand.Intn(40)) * time.Millisecond)

	return nil
}

// bezierPath computes a cubic Bézier curve between two points using random
// control points to introduce curvature and non‑linearity.
func bezierPath(x0, y0, x3, y3 float64, steps int) [][2]float64 {
	// Control points are offset from the straight line by a random amount,
	// creating a gentle arc rather than a straight segment.
	cp1x := x0 + (x3-x0)/3 + (rand.Float64()-0.5)*60
	cp1y := y0 + (y3-y0)/3 + (rand.Float64()-0.5)*60
	cp2x := x0 + 2*(x3-x0)/3 + (rand.Float64()-0.5)*60
	cp2y := y0 + 2*(y3-y0)/3 + (rand.Float64()-0.5)*60

	path := make([][2]float64, 0, steps+1)
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		mt := 1 - t
		x := mt*mt*mt*x0 + 3*mt*mt*t*cp1x + 3*mt*t*t*cp2x + t*t*t*x3
		y := mt*mt*mt*y0 + 3*mt*mt*t*cp1y + 3*mt*t*t*cp2y + t*t*t*y3
		path = append(path, [2]float64{x, y})
	}
	return path
}



