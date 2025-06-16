package transitions

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type TransitionI interface {
	Update() bool
	Draw(screen *ebiten.Image)
}
