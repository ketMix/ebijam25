package world

import (
	"fmt"
	"math/rand"

	"github.com/KEINOS/go-noise"
)

type Fate struct {
	noise.Generator
	NumGen    *rand.Rand
	certainty float64
}

func NewFate(sneed uint) Fate {
	generator, err := noise.New(noise.OpenSimplex, int64(sneed))
	if err != nil {
		panic("failed to create fate with sneed: " + fmt.Sprint(sneed) + err.Error())
	}

	return Fate{
		Generator: generator,
		NumGen:    rand.New(rand.NewSource(int64(sneed))),
		certainty: 200,
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
