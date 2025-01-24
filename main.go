package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"mines/internal/mines"
)

func main() {
	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Minesweeper")
	if err := ebiten.RunGame(mines.NewGame()); err != nil {
		panic(err)
	}
}
