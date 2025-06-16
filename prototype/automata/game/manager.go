package game

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type DrawManagerI interface {
	Draw(screen *ebiten.Image)
}

type UpdateManagerI interface {
	Update()
}

type ManagerI interface{}

type Managers struct {
	updaters []UpdateManagerI
	drawers  []DrawManagerI
}

func (ms *Managers) Add(m ManagerI) {
	if m2, ok := m.(UpdateManagerI); ok {
		ms.updaters = append(ms.updaters, m2)
	}
	if m2, ok := m.(DrawManagerI); ok {
		ms.drawers = append(ms.drawers, m2)
	}
}

func (ms *Managers) Remove(m ManagerI) {
	for i, m2 := range ms.updaters {
		if m == m2 {
			ms.updaters = append(ms.updaters[:i], ms.updaters[i+1:]...)
			break
		}
	}
	for i, m2 := range ms.drawers {
		if m == m2 {
			ms.drawers = append(ms.drawers[:i], ms.drawers[i+1:]...)
			break
		}
	}
}

func (ms *Managers) Update() {
	for _, m := range ms.updaters {
		m.Update()
	}
}

func (ms *Managers) Draw(screen *ebiten.Image) {
	for _, m := range ms.drawers {
		m.Draw(screen)
	}
}
