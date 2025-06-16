package transitions

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Manager struct {
	transitions []TransitionI
}

func (m *Manager) Top() TransitionI {
	if len(m.transitions) == 0 {
		return nil
	}
	return m.transitions[0]
}

func (m *Manager) Update() {
	if top := m.Top(); top != nil && top.Update() {
		m.transitions = m.transitions[1:]
	}
}

func (m *Manager) Draw(screen *ebiten.Image) {
	if top := m.Top(); top != nil {
		top.Draw(screen)
	}
}

func (m *Manager) Add(t TransitionI) {
	m.transitions = append(m.transitions, t)
}
