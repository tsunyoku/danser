package dance

import (
	"github.com/tsunyoku/danser/app/beatmap"
	"github.com/tsunyoku/danser/app/beatmap/objects"
	"github.com/tsunyoku/danser/app/dance/movers"
	"github.com/tsunyoku/danser/app/dance/schedulers"
	"github.com/tsunyoku/danser/app/dance/spinners"
	"github.com/tsunyoku/danser/app/graphics"
	"github.com/tsunyoku/danser/app/settings"
	"strings"
)

type Controller interface {
	SetBeatMap(beatMap *beatmap.BeatMap)
	InitCursors()
	Update(time float64, delta float64)
	GetCursors() []*graphics.Cursor
}

type GenericController struct {
	bMap       *beatmap.BeatMap
	cursors    []*graphics.Cursor
	schedulers []schedulers.Scheduler
}

func NewGenericController() Controller {
	return &GenericController{}
}

func (controller *GenericController) SetBeatMap(beatMap *beatmap.BeatMap) {
	controller.bMap = beatMap
}

func (controller *GenericController) InitCursors() {
	controller.cursors = make([]*graphics.Cursor, settings.TAG)
	controller.schedulers = make([]schedulers.Scheduler, settings.TAG)

	// Mover initialization
	for i := range controller.cursors {
		controller.cursors[i] = graphics.NewCursor()

		mover := "flower"
		if len(settings.Dance.Movers) > 0 {
			mover = strings.ToLower(settings.Dance.Movers[i%len(settings.Dance.Movers)])
		}

		var scheduler schedulers.Scheduler

		switch mover {
		case "spline":
			scheduler = schedulers.NewGenericScheduler(movers.NewSplineMover)
		case "bezier":
			scheduler = schedulers.NewGenericScheduler(movers.NewBezierMover)
		case "circular":
			scheduler = schedulers.NewGenericScheduler(movers.NewHalfCircleMover)
		case "linear":
			scheduler = schedulers.NewGenericScheduler(movers.NewLinearMover)
		case "axis":
			scheduler = schedulers.NewGenericScheduler(movers.NewAxisMover)
		case "exgon":
			scheduler = schedulers.NewGenericScheduler(movers.NewExGonMover)
		case "aggressive":
			scheduler = schedulers.NewGenericScheduler(movers.NewAggressiveMover)
		case "momentum":
			scheduler = schedulers.NewGenericScheduler(movers.NewMomentumMover)
		default:
			scheduler = schedulers.NewGenericScheduler(movers.NewAngleOffsetMover)
		}

		controller.schedulers[i] = scheduler
	}

	type Queue struct {
		objs []objects.IHitObject
	}

	objs := make([]Queue, settings.TAG)

	queue := controller.bMap.GetObjectsCopy()

	// Convert retarded (0 length / 0ms) sliders to pseudo-circles
	for i := 0; i < len(queue); i++ {
		if s, ok := queue[i].(*objects.Slider); ok && s.IsRetarded() {
			queue = schedulers.PreprocessQueue(i, queue, true)
		}
	}

	// Convert sliders to pseudo-circles for tag cursors
	if !settings.Dance.Battle && settings.Dance.TAGSliderDance && settings.TAG > 1 {
		for i := 0; i < len(queue); i++ {
			queue = schedulers.PreprocessQueue(i, queue, true)
		}
	}

	// Resolving 2B conflicts
	for i := 0; i < len(queue); i++ {
		if s, ok := queue[i].(*objects.Slider); ok {
			for j := i + 1; j < len(queue); j++ {
				o := queue[j]
				if (o.GetStartTime() >= s.GetStartTime() && o.GetStartTime() <= s.GetEndTime()) || (o.GetEndTime() >= s.GetStartTime() && o.GetEndTime() <= s.GetEndTime()) {
					queue = schedulers.PreprocessQueue(i, queue, true)
					break
				}
			}
		}
	}

	// If DoSpinnersTogether is true with tag mode, allow all tag cursors to spin the same spinner with different movers
	for j, o := range queue {
		if _, ok := o.(*objects.Spinner); (ok && settings.Dance.DoSpinnersTogether) || settings.Dance.Battle {
			for i := range objs {
				objs[i].objs = append(objs[i].objs, o)
			}
		} else {
			i := j % settings.TAG
			objs[i].objs = append(objs[i].objs, o)
		}
	}

	//Initialize spinner movers
	for i := range controller.cursors {
		spinMover := "circle"
		if len(settings.Dance.Spinners) > 0 {
			spinMover = settings.Dance.Spinners[i%len(settings.Dance.Spinners)]
		}

		controller.schedulers[i].Init(objs[i].objs, controller.bMap.Diff.Mods, controller.cursors[i], spinners.GetMoverCtorByName(spinMover), true)
	}
}

func (controller *GenericController) Update(time float64, delta float64) {
	for i := range controller.cursors {
		controller.schedulers[i].Update(time)
		controller.cursors[i].Update(delta)

		controller.cursors[i].LeftButton = controller.cursors[i].LeftKey || controller.cursors[i].LeftMouse
		controller.cursors[i].RightButton = controller.cursors[i].RightKey || controller.cursors[i].RightMouse
	}
}

func (controller *GenericController) GetCursors() []*graphics.Cursor {
	return controller.cursors
}
