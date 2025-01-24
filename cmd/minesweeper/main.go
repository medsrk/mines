package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"mines/internal/game"
	"mines/internal/renderer"
)

func main() {
	g := game.NewMinesweeper(20, 20)
	r := renderer.NewEbitenRenderer(g, 32)

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Minesweeper")
	if err := ebiten.RunGame(r); err != nil {
		panic(err)
	}
}
