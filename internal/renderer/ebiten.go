package renderer

import (
	"fmt"
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

const headerHeight = 50

type EbitenRenderer struct {
	game      game.Game
	cellSize  int
	font      font.Face
	sprites   *Sprites
	transform ebiten.GeoM
}

func NewEbitenRenderer(g game.Game, cellSize int) *EbitenRenderer {
	sprites, err := LoadSprites()
	if err != nil {
		panic(fmt.Sprintf("failed to load sprites: %v", err))
	}

	var transform ebiten.GeoM
	transform.Translate(0, headerHeight)

	return &EbitenRenderer{
		game:      g,
		cellSize:  cellSize,
		font:      basicfont.Face7x13,
		sprites:   sprites,
		transform: transform,
	}
}

func (r *EbitenRenderer) Update() error {
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		// Transform screen coordinates back to game coordinates
		invTransform := r.transform
		invTransform.Invert()
		gx, gy := invTransform.Apply(float64(x), float64(y))

		pos := game.Position{
			X: int(gx) / r.cellSize,
			Y: int(gy) / r.cellSize,
		}
		if pos.X >= 0 && pos.X < r.game.Width() && pos.Y >= 0 && pos.Y < r.game.Height() {
			r.game.HandleLeftClick(pos)
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		x, y := ebiten.CursorPosition()
		// Transform screen coordinates back to game coordinates
		invTransform := r.transform
		invTransform.Invert()
		gx, gy := invTransform.Apply(float64(x), float64(y))

		pos := game.Position{
			X: int(gx) / r.cellSize,
			Y: int(gy) / r.cellSize,
		}
		if pos.X >= 0 && pos.X < r.game.Width() && pos.Y >= 0 && pos.Y < r.game.Height() {
			r.game.HandleRightClick(pos)
		}
	}

	r.game.Update()
	return nil
}

func (r *EbitenRenderer) Draw(screen *ebiten.Image) {
	width := r.game.Width() * r.cellSize

	// Draw header directly to screen (no transform)
	vector.DrawFilledRect(screen, 0, 0, float32(width), headerHeight, color.RGBA{40, 40, 40, 255}, true)

	state := r.game.GameState()
	minesLeft := r.game.(*game.Minesweeper).MineCount() - r.game.FlagCount()

	minutes := int(state.ElapsedTime.Minutes())
	seconds := int(state.ElapsedTime.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", minutes, seconds)

	text.Draw(screen, fmt.Sprintf("Mines: %d", minesLeft), r.font, 20, headerHeight/2+6, color.White)
	bounds, _ := font.BoundString(r.font, timeStr)
	timeWidth := (bounds.Max.X - bounds.Min.X).Ceil()
	text.Draw(screen, timeStr, r.font, width-timeWidth-20, headerHeight/2+6, color.White)

	// draw game over or win text in center of header
	if state.IsGameOver {
		var textStr string
		var col color.Color
		if state.HasWon {
			textStr = "You Win!"
			col = color.RGBA{0, 255, 0, 255}
		} else {
			textStr = "Game Over"
			col = color.RGBA{255, 0, 0, 255}
		}
		bounds, _ := font.BoundString(r.font, textStr)
		textWidth := (bounds.Max.X - bounds.Min.X).Ceil()
		text.Draw(screen, textStr, r.font, width/2-textWidth/2, headerHeight/2+6, col)
	}

	// Draw cells with offset
	for x := 0; x < r.game.Width(); x++ {
		for y := 0; y < r.game.Height(); y++ {
			pos := game.Position{X: x, Y: y}
			r.drawCell(screen, pos, headerHeight)
		}
	}
}

func (r *EbitenRenderer) Layout(w, h int) (int, int) {
	return r.game.Width() * r.cellSize,
		r.game.Height()*r.cellSize + headerHeight
}

func (r *EbitenRenderer) drawCell(screen *ebiten.Image, pos game.Position, yOffset int) {
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
	y := float32(pos.Y*r.cellSize) + float32(yOffset)
	size := float32(r.cellSize)

	centerX := x + size/2
	centerY := y + size/2
	x = centerX - (size*scale)/2
	y = centerY - (size*scale)/2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(x), float64(y))

	switch state {
	case game.StateHidden, game.StateRevealing:
		screen.DrawImage(r.sprites.hidden, op)
	case game.StateRevealed:
		screen.DrawImage(r.sprites.revealed, op)
		if r.game.GetCellContent(pos) == game.ContentMine {
			screen.DrawImage(r.sprites.mine, op)
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
		screen.DrawImage(r.sprites.hidden, op)
		screen.DrawImage(r.sprites.flag, op)
	}
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
