package main

import "github.com/hajimehoshi/ebiten/v2"

type Game struct{}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
}

func (g *Game) Layout(ow, oh int) (int, int) {
	return ow, oh
}

func main() {
	g := &Game{}
	ebiten.SetWindowSize(600, 600)
	ebiten.SetWindowTitle("Mobbox")
	if err := ebiten.RunGame(g); err != nil {
		panic(err)
	}
}
