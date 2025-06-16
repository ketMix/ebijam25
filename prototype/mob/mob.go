package main

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/kettek/rebui"
	"github.com/kettek/rebui/widgets"
)

type Mob struct {
	id               int
	participants     []Participant // an individual or structure or whatever.
	x, y             float64
	targetX, targetY float64
	targetId         int
	debugNode        *rebui.Node
}

var mobIdCounter = 0

func nextMobId() int {
	mobIdCounter++
	return mobIdCounter
}

func NewMob(x, y float64) *Mob {
	node := gLayout.AddNode(rebui.Node{
		Type:            "Text",
		Width:           "128",
		Height:          "32",
		X:               "20",
		Y:               "20",
		Text:            "Mob",
		ForegroundColor: "white",
	})
	node.Widget.(*widgets.Text).AssignBorderColor(nil)
	return &Mob{
		id:           nextMobId(),
		x:            x,
		y:            y,
		targetX:      x,
		targetY:      y,
		participants: make([]Participant, 0),
		debugNode:    node,
	}
}

func (m *Mob) Radius() float64 {
	return float64(len(m.participants)) * 2
}

func (m *Mob) Individuals() []*Individual {
	var individuals []*Individual
	for _, p := range m.participants {
		if ind, ok := p.(*Individual); ok {
			individuals = append(individuals, ind)
		}
	}
	return individuals
}

func (m *Mob) Structures() []*Structure {
	var structures []*Structure
	for _, p := range m.participants {
		if str, ok := p.(*Structure); ok {
			structures = append(structures, str)
		}
	}
	return structures
}

func (m *Mob) AddParticipant(p Participant) {
	m.participants = append(m.participants, p)
}

func (m *Mob) RemoveIndividuals(count int) {
	for i := len(m.participants) - 1; i >= 0 && count > 0; i-- {
		if _, ok := m.participants[i].(*Individual); ok {
			m.participants = append(m.participants[:i], m.participants[i+1:]...)
			count--
		}
	}
}

func (m *Mob) Draw(screen *ebiten.Image) {
	radius := len(m.participants) * 2
	vector.StrokeCircle(screen, float32(m.x), float32(m.y), float32(radius)/2, 1, color.NRGBA{255, 0, 255, 255}, true)

	for _, p := range m.participants {
		p.Draw(screen)
	}
}

func (m *Mob) Update(g *Game) {
	m.debugNode.Widget.(*widgets.Text).AssignText(fmt.Sprintf("%d: %d participants", m.id, len(m.participants)))
	m.debugNode.Widget.(*widgets.Text).AssignX(m.x)
	m.debugNode.Widget.(*widgets.Text).AssignY(m.y)
	if m.targetId != 0 {
		var targetMob *Mob
		for _, mob := range g.mobs {
			if mob.id == m.targetId {
				targetMob = mob
				break
			}
		}
		if targetMob != nil {
			m.targetX = targetMob.x
			m.targetY = targetMob.y
		} else {
			// If the target mob is not found, reset the target.
			m.targetId = 0
			m.targetX = m.x
			m.targetY = m.y
		}
	}

	// See if we can merge mobs.
	pullX := 0.0
	pullY := 0.0
	for _, mob := range g.mobs {
		if mob.id == m.id {
			continue
		}
		if len(m.participants) >= len(mob.participants) {
			// Circle check.
			dx := m.x - mob.x
			dy := m.y - mob.y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 4 || dist < (m.Radius()+mob.Radius())/3 {
				eventBus.Publish(&EventMerge{
					from: mob.id,
					to:   m.id,
				})
				// Just stop processing for now.
				return
			}
		}
		// Pull towards the other mob.
		if len(m.participants) <= len(mob.participants) {
			dx := mob.x - m.x
			dy := mob.y - m.y
			dist := math.Sqrt(dx*dx + dy*dy)
			fmt.Println(mob.Radius()*2, m.Radius(), dist)
			sizeDiff := math.Min((mob.Radius()*2 - m.Radius()), 3)
			if dist < 500 { // Adjust this threshold as needed.
				// Pull towards the other mob.
				pullX += dx / dist * sizeDiff // Adjust speed as needed.
				pullY += dy / dist * sizeDiff
			}
		}
	}
	if pullX != 0 || pullY != 0 {
		// Move towards the other mob.
		m.x += pullX
		m.y += pullY
	}

	// Move towards targetX and targetY.
	if m.x != m.targetX || m.y != m.targetY {
		angleToTarget := math.Atan2(m.targetY-m.y, m.targetX-m.x)
		dx := math.Cos(angleToTarget) * 2
		dy := math.Sin(angleToTarget) * 2
		m.x += dx
		m.y += dy
		// If close enough, snap to target.
		if math.Abs(m.x-m.targetX) < 1 && math.Abs(m.y-m.targetY) < 1 {
			m.x = m.targetX
			m.y = m.targetY
		}
	}

	for _, p := range m.participants {
		p.Update(m.participants)
		// Also pull them towards our position, greater the further away they are.
		if ind, ok := p.(*Individual); ok {
			dx := m.x - ind.x
			dy := m.y - ind.y
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist > 0 {
				// Move towards the mob's position.
				ind.x += dx / dist * 1.5 // Adjust speed as needed.
				ind.y += dy / dist * 1.5
			}
		}
	}
}
