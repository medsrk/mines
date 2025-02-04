package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"mines/internal/game"
	"mines/internal/renderer"
)

func main() {
	g := game.NewMinesweeper(10, 10, 10)
	r := renderer.NewEbitenRenderer(g, 32)

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Minesweeper")
	if err := ebiten.RunGame(r); err != nil {
		panic(err)
	}
}
