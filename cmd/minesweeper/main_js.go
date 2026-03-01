//go:build js

package main

import (
	"syscall/js"

	"github.com/hajimehoshi/ebiten/v2"
	"mines/internal/renderer"
)

func main() {
	w := js.Global().Get("window").Get("innerWidth").Int()
	h := js.Global().Get("window").Get("innerHeight").Int()

	r := renderer.NewEbitenRenderer(nil, 32)

	ebiten.SetWindowSize(w, h)
	ebiten.SetWindowTitle("Minesweeper")
	if err := ebiten.RunGame(r); err != nil {
		panic(err)
	}
}
