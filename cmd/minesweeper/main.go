package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"mines/internal/renderer"
)

func main() {
	r := renderer.NewEbitenRenderer(nil, 32)

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Minesweeper")
	if err := ebiten.RunGame(r); err != nil {
		panic(err)
	}
}
