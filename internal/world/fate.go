package world

import (
	"fmt"

	"github.com/KEINOS/go-noise"
)

type Fate struct {
	noise.Generator
	certainty float64
}

func NewFate(sneed uint) Fate {
	generator, err := noise.New(noise.OpenSimplex, int64(sneed))
	if err != nil {
		panic("failed to create fate with sneed: " + fmt.Sprint(sneed) + err.Error())
	}

	return Fate{
		Generator: generator,
		certainty: 0.5, // Default smoothness
	}
}

func (f *Fate) Determine(values ...float64) float64 {
	if len(values) == 0 {
		return f.Eval64(0.0)
	}

	smoothed := make([]float64, len(values))
	for i, v := range values {
		smoothed[i] = v / f.certainty
	}
	return f.Eval64(smoothed...)
}
