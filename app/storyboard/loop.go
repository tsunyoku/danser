package storyboard

import (
	"github.com/tsunyoku/danser/framework/math/animation"
	"log"
	"math"
	"strconv"
)

type LoopProcessor struct {
	start, repeats int64
	transforms     []*animation.Transformation
}

func NewLoopProcessor(data []string) *LoopProcessor {
	loop := new(LoopProcessor)

	var err error

	loop.start, err = strconv.ParseInt(data[1], 10, 64)
	if err != nil {
		log.Println("Failed to parse: ", data)
		panic(err)
	}

	loop.repeats, err = strconv.ParseInt(data[2], 10, 64)
	if err != nil {
		log.Println("Failed to parse: ", data)
		panic(err)
	}

	return loop
}

func (loop *LoopProcessor) Add(command []string) {
	loop.transforms = append(loop.transforms, parseCommand(command)...)
}

func (loop *LoopProcessor) Unwind() []*animation.Transformation {
	var transforms []*animation.Transformation

	startTime := math.MaxFloat64
	endTime := -math.MaxFloat64

	for _, t := range loop.transforms {
		startTime = math.Min(startTime, t.GetStartTime())
		endTime = math.Max(endTime, t.GetEndTime())
	}

	iterationTime := endTime - startTime

	for i := int64(0); i < loop.repeats; i++ {
		partStart := float64(loop.start) + float64(i)*iterationTime

		for _, t := range loop.transforms {
			transforms = append(transforms, t.Clone(partStart+t.GetStartTime(), partStart+t.GetEndTime()))
		}
	}

	return transforms
}
