package renderer

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"image/color"
	"mines/internal/game"
	"strconv"
	"time"
)

type EbitenRenderer struct {
	game     game.Game
	cellSize int
	font     font.Face
}

func NewEbitenRenderer(g game.Game, cellSize int) *EbitenRenderer {
	return &EbitenRenderer{
		game:     g,
		cellSize: cellSize,
		font:     basicfont.Face7x13,
	}
}

func (r *EbitenRenderer) Update() error {
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		pos := game.Position{
			X: x / r.cellSize,
			Y: y / r.cellSize,
		}
		if pos.X < r.game.Width() && pos.Y < r.game.Height() {
			r.game.HandleLeftClick(pos)
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		x, y := ebiten.CursorPosition()
		pos := game.Position{
			X: x / r.cellSize,
			Y: y / r.cellSize,
		}
		if pos.X < r.game.Width() && pos.Y < r.game.Height() {
			r.game.HandleRightClick(pos)
		}
	}

	r.game.Update()

	return nil
}

func (r *EbitenRenderer) Draw(screen *ebiten.Image) {
	for x := 0; x < r.game.Width(); x++ {
		for y := 0; y < r.game.Height(); y++ {
			pos := game.Position{X: x, Y: y}
			r.drawCell(screen, pos)
		}
	}
}

func (r *EbitenRenderer) Layout(w, h int) (int, int) {
	return r.game.Width() * r.cellSize,
		r.game.Height() * r.cellSize
}

func (r *EbitenRenderer) drawCell(screen *ebiten.Image, pos game.Position) {
	scale := float32(1.0)
	state := r.game.GetCellState(pos)

	if state == game.StateRevealing {
		if anim, exists := r.game.Grid().GetAnimation(pos); exists {
			elapsed := time.Since(anim.StartTime)
			if elapsed >= 0 {
				progress := float64(elapsed) / float64(anim.Duration)
				if progress < 1.0 {
					scale = float32(0.5 + 0.5*progress)
				}
			}
		}
	}

	x := float32(pos.X * r.cellSize)
	y := float32(pos.Y * r.cellSize)
	size := float32(r.cellSize)

	centerX := x + size/2
	centerY := y + size/2
	x = centerX - (size*scale)/2
	y = centerY - (size*scale)/2
	size *= scale

	switch state {
	case game.StateHidden, game.StateRevealing:
		vector.DrawFilledRect(screen, x, y, size, size, color.RGBA{127, 127, 127, 255}, true)
	case game.StateRevealed:
		vector.DrawFilledRect(screen, x, y, size, size, color.RGBA{61, 61, 61, 255}, true)
		if r.game.GetCellContent(pos) == game.ContentMine {
			vector.DrawFilledRect(screen, x, y, size, size, color.RGBA{255, 0, 0, 255}, true)
		} else {
			mineCount := r.game.GetAdjacentMines(pos)
			if mineCount > 0 {
				numberStr := strconv.Itoa(mineCount)
				bounds, _ := font.BoundString(r.font, numberStr)
				width := (bounds.Max.X - bounds.Min.X).Ceil()
				height := (bounds.Max.Y - bounds.Min.Y).Ceil()
				textX := int(x) + (r.cellSize-width)/2
				textY := int(y) + (r.cellSize+height)/2
				text.Draw(screen, numberStr, r.font, textX, textY, numberColours[mineCount])
			}
		}
	case game.StateFlagged:
		vector.DrawFilledRect(screen, x, y, size, size, color.RGBA{0, 0, 255, 255}, true)
	case game.StateQuestion:
		vector.DrawFilledRect(screen, x, y, size, size, color.RGBA{255, 255, 0, 255}, true)
	}
	vector.StrokeRect(screen, x, y, size, size, 1, color.Black, true)
}

var numberColours = []color.Color{
	color.RGBA{128, 128, 128, 255}, // 0 - grey
	color.RGBA{255, 255, 255, 255}, // 1 - white
	color.RGBA{118, 253, 75, 255},  // 2 - green
	color.RGBA{255, 255, 85, 255},  // 3 - yellow
	color.RGBA{255, 0, 0, 255},     // 4 - red
	color.RGBA{128, 0, 0, 255},     // 5 - dark red
	color.RGBA{0, 128, 128, 255},   // 6 - teal
	color.RGBA{0, 0, 0, 255},       // 7 - black
	color.RGBA{128, 128, 128, 255}, // 8 - grey
}
