package spinners

import (
	"github.com/tsunyoku/danser/app/settings"
	"github.com/tsunyoku/danser/framework/math/math32"
	"github.com/tsunyoku/danser/framework/math/vector"
)

type CircleMover struct {
	start float64
}

func NewCircleMover() *CircleMover {
	return &CircleMover{}
}

func (c *CircleMover) Init(start, end float64) {
	c.start = start
}

func (c *CircleMover) GetPositionAt(time float64) vector.Vector2f {
	return vector.NewVec2fRad(rpms*float32(time-c.start)*2*math32.Pi, float32(settings.Dance.SpinnerRadius)).Add(center)
}
