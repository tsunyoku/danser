package curves

import (
	"github.com/tsunyoku/danser/framework/math/math32"
	"github.com/tsunyoku/danser/framework/math/vector"
	"math"
)

type CirArc struct {
	pt1, pt2, pt3               vector.Vector2f
	centre                      vector.Vector2f //nolint:misspell
	startAngle, totalAngle, dir float64
	r                           float32
	Unstable                    bool
}

func NewCirArc(a, b, c vector.Vector2f) *CirArc {
	arc := &CirArc{pt1: a, pt2: b, pt3: c, dir: 1}

	if math32.Abs((b.Y-a.Y)*(c.X-a.X)-(b.X-a.X)*(c.Y-a.Y)) < 0.001 {
		arc.Unstable = true
	}

	d := 2 * (a.X*(b.Y-c.Y) + b.X*(c.Y-a.Y) + c.X*(a.Y-b.Y))
	aSq := a.LenSq()
	bSq := b.LenSq()
	cSq := c.LenSq()

	arc.centre = vector.NewVec2f(
		aSq*(b.Y-c.Y)+bSq*(c.Y-a.Y)+cSq*(a.Y-b.Y),
		aSq*(c.X-b.X)+bSq*(a.X-c.X)+cSq*(b.X-a.X)).Scl(1 / d) //nolint:misspell

	arc.r = a.Dst(arc.centre)
	arc.startAngle = math.Atan2(float64(a.Y)-float64(arc.centre.Y), float64(a.X)-float64(arc.centre.X))

	endAngle := math.Atan2(float64(c.Y)-float64(arc.centre.Y), float64(c.X)-float64(arc.centre.X))

	for endAngle < arc.startAngle {
		endAngle += 2 * math.Pi
	}

	arc.totalAngle = endAngle - arc.startAngle

	aToC := c.Sub(a)
	aToC = vector.NewVec2f(aToC.Y, -aToC.X)

	if aToC.Dot(b.Sub(a)) < 0 {
		arc.dir = -arc.dir
		arc.totalAngle = 2*math.Pi - arc.totalAngle
	}

	return arc
}

func (arc *CirArc) PointAt(t float32) vector.Vector2f {
	return vector.NewVec2dRad(arc.startAngle+arc.dir*float64(t)*arc.totalAngle, float64(arc.r)).Copy32().Add(arc.centre)
}

func (arc *CirArc) GetLength() float32 {
	return float32(float64(arc.r) * arc.totalAngle)
}

func (arc *CirArc) GetStartAngle() float32 {
	return arc.pt1.AngleRV(arc.PointAt(1.0 / arc.GetLength()))
}

func (arc *CirArc) GetEndAngle() float32 {
	return arc.pt3.AngleRV(arc.PointAt((arc.GetLength() - 1.0) / arc.GetLength()))
}
