package transitions

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Fade struct {
	lifetime float64
	duration float64
	fadeIn   bool
}

func NewFade(duration int, fadeIn bool) *Fade {
	return &Fade{
		duration: float64(duration),
		fadeIn:   fadeIn,
	}
}

func (f *Fade) Update() bool {
	f.lifetime++
	return f.lifetime >= f.duration
}

func (f *Fade) Draw(screen *ebiten.Image) {
	clr := color.NRGBA{}
	if f.fadeIn {
		clr.A = uint8(f.lifetime / f.duration * 255)
	} else {
		clr.A = uint8(255 - f.lifetime/f.duration*255)
	}
	vector.DrawFilledRect(screen, 0, 0, float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()), clr, false)
}
