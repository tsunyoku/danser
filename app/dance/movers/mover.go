package movers

import (
	"github.com/tsunyoku/danser/app/beatmap/difficulty"
	"github.com/tsunyoku/danser/app/beatmap/objects"
	"github.com/tsunyoku/danser/framework/math/vector"
)

type MultiPointMover interface {
	Reset(mods difficulty.Modifier)
	SetObjects(objs []objects.IHitObject) int
	Update(time float64) vector.Vector2f
	GetEndTime() float64
}
