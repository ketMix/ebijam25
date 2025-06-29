package client

import (
	"fmt"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/ketMix/ebijam25/internal/world"
	"github.com/kettek/rebui"
	"github.com/kettek/rebui/widgets"
)

type Hiscore struct {
	layout  rebui.Layout
	entries []*rebui.Node
}

func (h *Hiscore) Update(players []*world.Player) {
	// Clone and sort by player.count.
	var sortedPlayers []*world.Player
	sortedPlayers = make([]*world.Player, len(players))
	copy(sortedPlayers, players)
	slices.SortFunc(sortedPlayers, func(a, b *world.Player) int {
		return b.Count - a.Count
	})
	// First see if we can reuse existing entries
	for i, player := range sortedPlayers {
		if i < len(h.entries) {
			h.entries[i].Widget.(*widgets.Text).AssignText(player.Username + " - " + fmt.Sprintf("%d", player.Count))
			h.entries[i].Widget.(*widgets.Text).AssignForegroundColor(player.Color)
		} else {
			var y string
			if i > 0 {
				y = "after " + h.entries[i-1].ID
			} else {
				y = "0%"
			}
			entry := h.layout.AddNode(rebui.Node{
				Type:            "Text",
				ID:              "entry-" + fmt.Sprintf("%d", i),
				Width:           "30%",
				Height:          "40%",
				X:               "100%",
				OriginX:         "-100%",
				Y:               y,
				Text:            player.Username + " - " + fmt.Sprintf("%d", player.Count),
				VerticalAlign:   rebui.AlignTop,
				HorizontalAlign: rebui.AlignRight,
			})
			entry.Widget.(*widgets.Text).AssignBackgroundColor(nil)
			entry.Widget.(*widgets.Text).AssignBorderColor(nil)
			h.entries = append(h.entries, entry)
		}
	}
	// Remove any excess entries
	for i := len(players); i < len(h.entries); i++ {
		h.layout.RemoveNode(h.entries[i])
	}
	h.entries = h.entries[:len(players)]
}

func (h *Hiscore) Draw(screen *ebiten.Image) {
	h.layout.Draw(screen)
}

func (h *Hiscore) Layout(width, height int) {
	h.layout.Layout(float64(width), float64(height))
}
