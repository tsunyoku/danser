package schedulers

import (
	"github.com/tsunyoku/danser/app/beatmap/difficulty"
	"github.com/tsunyoku/danser/app/beatmap/objects"
	"github.com/tsunyoku/danser/app/dance/spinners"
	"github.com/tsunyoku/danser/app/graphics"
)

type Scheduler interface {
	Init(objects []objects.IHitObject, mods difficulty.Modifier, cursor *graphics.Cursor, spinnerMoverCtor func() spinners.SpinnerMover, initKeys bool)
	Update(time float64)
}
