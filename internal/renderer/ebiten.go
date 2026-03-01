package renderer

import (
	"fmt"
	"image/color"
	"math"
	"mines/internal/game"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

const (
	headerHeight = 50
	gameOverText = "Game Over"
	winText      = "You Win!"
	menuText     = "Minesweeper"
)

var (
	gameOverColor = color.RGBA{255, 0, 0, 255}
	winColor      = color.RGBA{0, 255, 0, 255}
)

type MenuState int

const (
	MenuStateMain MenuState = iota
	MenuStatePlaying
)

type Difficulty struct {
	Name      string
	Width     int
	Height    int
	MineCount int
}

var difficulties = []Difficulty{
	{Name: "Easy", Width: 10, Height: 10, MineCount: 10},
	{Name: "Medium", Width: 16, Height: 16, MineCount: 40},
	{Name: "Hard", Width: 24, Height: 24, MineCount: 99},
}

const (
	revealDelay   = 60 * time.Millisecond
	revealAnimate = 20 * time.Millisecond
)

type cellAnimation struct {
	StartTime time.Time
	Duration  time.Duration
}

const longPressThreshold = 30 // ticks (0.5s at 60fps)

type touchInfo struct {
	pos      game.Position
	consumed bool
}

type EbitenRenderer struct {
	game      game.Game
	cellSize  int
	font      font.Face
	sprites   *Sprites
	audio     *Audio
	transform ebiten.GeoM
	menuState MenuState
	menuItems []string

	animations map[game.Position]cellAnimation
	touches    map[ebiten.TouchID]touchInfo

	screenW int
	screenH int

	customWidth  int
	customHeight int
	customMines  int
}

func NewEbitenRenderer(g game.Game, cellSize int) *EbitenRenderer {
	sprites, err := loadSprites()
	if err != nil {
		panic(fmt.Sprintf("failed to load sprites: %v", err))
	}

	var transform ebiten.GeoM
	transform.Translate(0, headerHeight)

	menuItems := make([]string, len(difficulties))
	for i, d := range difficulties {
		menuItems[i] = fmt.Sprintf("%s (%dx%d, %d mines)", d.Name, d.Width, d.Height, d.MineCount)
	}

	return &EbitenRenderer{
		game:       g,
		cellSize:   cellSize,
		font:       basicfont.Face7x13,
		sprites:    sprites,
		audio:      loadAudio(),
		transform:  transform,
		menuState:  MenuStateMain,
		menuItems:  menuItems,
		animations: make(map[game.Position]cellAnimation),
		touches:    make(map[ebiten.TouchID]touchInfo),

		screenW: 1280,
		screenH: 720,

		customWidth:  10,
		customHeight: 10,
		customMines:  10,
	}
}

func (r *EbitenRenderer) Update() error {
	if r.menuState == MenuStateMain {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()
			cx := r.screenW / 2
			sy := float64(r.screenH) / 720.0

			// Handle preset difficulty selection
			for i, item := range r.menuItems {
				bounds, _ := font.BoundString(r.font, item)
				width := (bounds.Max.X - bounds.Min.X).Ceil()
				itemX := cx - width/2
				itemY := int(float64(200+i*50) * sy)
				if x >= itemX && x <= itemX+width && y >= itemY-20 && y <= itemY+20 {
					d := difficulties[i]
					r.game = game.NewMinesweeper(d.Width, d.Height, d.MineCount)
					r.menuState = MenuStatePlaying
					ebiten.SetWindowSize(d.Width*r.cellSize, d.Height*r.cellSize+headerHeight)
					winWidth, winHeight := ebiten.Monitor().Size()
					ebiten.SetWindowPosition((winWidth-d.Width*r.cellSize)/2, (winHeight-d.Height*r.cellSize-headerHeight)/2)
					return nil
				}
			}

			// Handle custom settings buttons
			baseY := int(float64(200+len(difficulties)*50+30) * sy)
			rowSpacing := int(float64(menuRowSpacing) * sy)
			rows := []struct {
				valuePtr *int
				minValue int
				maxValue int
			}{
				{valuePtr: &r.customWidth, minValue: 5, maxValue: 50},
				{valuePtr: &r.customHeight, minValue: 5, maxValue: 50},
				{valuePtr: &r.customMines, minValue: 1, maxValue: r.customWidth*r.customHeight - 1},
			}

			for i := range rows {
				field := []struct {
					label string
					value int
				}{
					{"Width:", r.customWidth}, {"Height:", r.customHeight}, {"Mines:", r.customMines},
				}[i]
				yPos := baseY + i*rowSpacing
				valueStr := fmt.Sprintf("%d", *rows[i].valuePtr)
				bounds, _ := font.BoundString(r.font, field.label)
				labelWidth := (bounds.Max.X - bounds.Min.X).Ceil()
				bounds, _ = font.BoundString(r.font, valueStr)
				valueWidth := (bounds.Max.X - bounds.Min.X).Ceil()
				totalWidth := labelWidth + menuButtonPadding + valueWidth + menuButtonPadding + menuButtonSize*2 + menuButtonPadding
				labelX := cx - totalWidth/2
				minusX := labelX + labelWidth + menuButtonPadding + valueWidth + menuButtonPadding
				plusX := minusX + menuButtonSize + menuButtonPadding

				// Minus button
				if x >= minusX && x <= minusX+menuButtonSize && y >= yPos && y <= yPos+menuButtonSize {
					if *rows[i].valuePtr > rows[i].minValue {
						*rows[i].valuePtr--
					}
				}
				// Plus button
				if x >= plusX && x <= plusX+menuButtonSize && y >= yPos && y <= yPos+menuButtonSize {
					if *rows[i].valuePtr < rows[i].maxValue {
						*rows[i].valuePtr++
					}
				}
			}

			// Start button
			startY := baseY + 3*rowSpacing
			startText := "Start Custom Game"
			bounds, _ := font.BoundString(r.font, startText)
			startWidth := (bounds.Max.X - bounds.Min.X).Ceil()
			startX := cx - startWidth/2
			if x >= startX-10 && x <= startX+startWidth+10 && y >= startY && y <= startY+30 {
				r.game = game.NewMinesweeper(r.customWidth, r.customHeight, r.customMines)
				r.menuState = MenuStatePlaying
				ebiten.SetWindowSize(r.customWidth*r.cellSize, r.customHeight*r.cellSize+headerHeight)
				winWidth, winHeight := ebiten.Monitor().Size()
				ebiten.SetWindowPosition((winWidth-r.customWidth*r.cellSize)/2, (winHeight-r.customHeight*r.cellSize-headerHeight)/2)
				return nil
			}
		}
		return nil
	}

	if r.game == nil {
		return nil
	}

	// --- Input Handling and Initial Action Sounds ---
	if !r.game.GameState().IsGameOver {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			if pos, clicked := r.mouseButtonClicked(); clicked {
				for _, event := range r.game.HandleLeftClick(pos) {
					switch event.Type {
					case game.EventExplosion:
						if r.audio.explodeSound != nil {
							r.audio.explodeSound.Rewind()
							r.audio.explodeSound.Play()
						}
					case game.EventReveal:
						if r.audio.initialRevealSound != nil {
							r.audio.initialRevealSound.Rewind()
							r.audio.initialRevealSound.Play()
						}
						for _, revealedPos := range event.Positions {
							dist := math.Abs(float64(revealedPos.X-pos.X)) + math.Abs(float64(revealedPos.Y-pos.Y))
							delay := time.Duration(dist) * revealDelay
							r.animations[revealedPos] = cellAnimation{
								StartTime: time.Now().Add(delay),
								Duration:  revealAnimate,
							}
						}
					}
				}
			}
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			if pos, clicked := r.mouseButtonClicked(); clicked {
				for _, event := range r.game.HandleRightClick(pos) {
					switch event.Type {
					case game.EventFlagged:
						if r.audio.flagSound != nil {
							r.audio.flagSound.Rewind()
							r.audio.flagSound.Play()
						}

					case game.EventUnflagged:
						if r.audio.unflagSound != nil {
							r.audio.unflagSound.Rewind()
							r.audio.unflagSound.Play()
						}
					}
				}
			}
		}

		// Touch: record new touches
		for _, id := range inpututil.JustPressedTouchIDs() {
			x, y := ebiten.TouchPosition(id)
			if pos, ok := r.screenToGamePos(x, y); ok {
				r.touches[id] = touchInfo{pos: pos}
			}
		}
		// Touch: long press → flag
		for id, t := range r.touches {
			if !t.consumed && inpututil.TouchPressDuration(id) >= longPressThreshold {
				t.consumed = true
				r.touches[id] = t
				for _, event := range r.game.HandleRightClick(t.pos) {
					switch event.Type {
					case game.EventFlagged:
						if r.audio.flagSound != nil {
							r.audio.flagSound.Rewind()
							r.audio.flagSound.Play()
						}
					case game.EventUnflagged:
						if r.audio.unflagSound != nil {
							r.audio.unflagSound.Rewind()
							r.audio.unflagSound.Play()
						}
					}
				}
			}
		}
		// Touch: release → reveal
		for _, id := range inpututil.AppendJustReleasedTouchIDs(nil) {
			if t, ok := r.touches[id]; ok {
				if !t.consumed {
					for _, event := range r.game.HandleLeftClick(t.pos) {
						switch event.Type {
						case game.EventExplosion:
							if r.audio.explodeSound != nil {
								r.audio.explodeSound.Rewind()
								r.audio.explodeSound.Play()
							}
						case game.EventReveal:
							if r.audio.initialRevealSound != nil {
								r.audio.initialRevealSound.Rewind()
								r.audio.initialRevealSound.Play()
							}
							for _, revealedPos := range event.Positions {
								dist := math.Abs(float64(revealedPos.X-t.pos.X)) + math.Abs(float64(revealedPos.Y-t.pos.Y))
								delay := time.Duration(dist) * revealDelay
								r.animations[revealedPos] = cellAnimation{
									StartTime: time.Now().Add(delay),
									Duration:  revealAnimate,
								}
							}
						}
					}
				}
				delete(r.touches, id)
			}
		}
	}
	for _, event := range r.game.Update() {
		switch event.Type {
		case game.EventWon:
			if r.audio.winSound != nil {
				r.audio.winSound.Rewind()
				r.audio.winSound.Play()
			}
		}
	}
	now := time.Now()
	for p, anim := range r.animations {
		if now.Sub(anim.StartTime) >= anim.Duration {
			delete(r.animations, p)
			if r.audio.tileRevealSound != nil {
				r.audio.tileRevealSound.Rewind()
				r.audio.tileRevealSound.Play()
			}
		}
	}

	// --- Reset Game ---
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		r.menuState = MenuStateMain
		ebiten.SetWindowSize(1280, 720)
		winWidth, winHeight := ebiten.Monitor().Size()
		ebiten.SetWindowPosition((winWidth-1280)/2, (winHeight-720)/2)
		r.animations = make(map[game.Position]cellAnimation)
		r.game = nil
	}

	return nil
}

func (r *EbitenRenderer) Draw(screen *ebiten.Image) {
	if r.menuState == MenuStateMain {
		r.drawMenu(screen)
		return
	}

	if r.game == nil {
		return
	}

	r.drawHeader(screen)

	for x := 0; x < r.game.Width(); x++ {
		for y := 0; y < r.game.Height(); y++ {
			pos := game.Position{X: x, Y: y}
			r.drawCell(screen, pos, headerHeight)
		}
	}
}

const (
	screenWidth       = 1280
	menuRowSpacing    = 50
	menuButtonSize    = 30
	menuButtonPadding = 10
)

func (r *EbitenRenderer) drawMenu(screen *ebiten.Image) {
	screen.Fill(color.RGBA{30, 30, 30, 255})
	cx := r.screenW / 2
	sy := float64(r.screenH) / 720.0

	// Draw menu title
	bounds, _ := font.BoundString(r.font, menuText)
	text.Draw(screen, menuText, r.font, cx-(bounds.Max.X-bounds.Min.X).Ceil()/2, int(100*sy), color.White)

	// Draw preset difficulty options
	for i, item := range r.menuItems {
		bounds, _ := font.BoundString(r.font, item)
		text.Draw(screen, item, r.font, cx-(bounds.Max.X-bounds.Min.X).Ceil()/2, int(float64(200+i*50)*sy), color.White)
	}

	// Calculate base Y dynamically
	baseY := int(float64(200+len(difficulties)*50+30) * sy)
	rowSpacing := int(float64(menuRowSpacing) * sy)
	for i, field := range []struct {
		label string
		value int
	}{
		{"Width:", r.customWidth}, {"Height:", r.customHeight}, {"Mines:", r.customMines},
	} {
		y := baseY + i*rowSpacing
		valueStr := fmt.Sprintf("%d", field.value)
		bounds, _ := font.BoundString(r.font, field.label)
		labelWidth := (bounds.Max.X - bounds.Min.X).Ceil()
		bounds, _ = font.BoundString(r.font, valueStr)
		valueWidth := (bounds.Max.X - bounds.Min.X).Ceil()
		totalWidth := labelWidth + menuButtonPadding + valueWidth + menuButtonPadding + menuButtonSize*2 + menuButtonPadding
		labelX := cx - totalWidth/2

		text.Draw(screen, field.label, r.font, labelX, y+15, color.White)
		text.Draw(screen, valueStr, r.font, labelX+labelWidth+menuButtonPadding, y+15, color.White)
		minusX := labelX + labelWidth + menuButtonPadding + valueWidth + menuButtonPadding
		vector.DrawFilledRect(screen, float32(minusX), float32(y), float32(menuButtonSize), float32(menuButtonSize), color.RGBA{100, 100, 100, 255}, false)
		text.Draw(screen, "-", r.font, minusX+(menuButtonSize-5)/2, y+15, color.White)
		plusX := minusX + menuButtonSize + menuButtonPadding
		vector.DrawFilledRect(screen, float32(plusX), float32(y), float32(menuButtonSize), float32(menuButtonSize), color.RGBA{100, 100, 100, 255}, false)
		text.Draw(screen, "+", r.font, plusX+(menuButtonSize-5)/2, y+15, color.White)
	}

	// Draw Start button
	startY := baseY + 3*rowSpacing
	startText := "Start Custom Game"
	bounds, _ = font.BoundString(r.font, startText)
	startWidth := (bounds.Max.X - bounds.Min.X).Ceil()
	startX := cx - startWidth/2
	vector.DrawFilledRect(screen, float32(startX)-10, float32(startY), float32(startWidth)+20, 30, color.RGBA{50, 150, 50, 255}, false)
	text.Draw(screen, startText, r.font, startX, startY+20, color.White)
}

func (r *EbitenRenderer) Layout(w, h int) (int, int) {
	r.screenW = w
	r.screenH = h
	if r.menuState == MenuStateMain || r.game == nil {
		return w, h
	}
	return r.game.Width() * r.cellSize,
		r.game.Height()*r.cellSize + headerHeight
}

func (r *EbitenRenderer) drawHeader(screen *ebiten.Image) {
	if r.game == nil {
		return
	}
	width := r.game.Width() * r.cellSize

	vector.DrawFilledRect(screen, 0, 0, float32(width), headerHeight, color.RGBA{40, 40, 40, 255}, true)

	state := r.game.GameState()
	minesLeft := r.game.MineCount() - r.game.FlagCount()

	minutes := int(state.ElapsedTime.Minutes())
	seconds := int(state.ElapsedTime.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d", minutes, seconds)

	text.Draw(screen, fmt.Sprintf("Mines: %d", minesLeft), r.font, 20, headerHeight/2+6, color.White)
	bounds, _ := font.BoundString(r.font, timeStr)
	timeWidth := (bounds.Max.X - bounds.Min.X).Ceil()
	text.Draw(screen, timeStr, r.font, width-timeWidth-20, headerHeight/2+6, color.White)

	if state.IsGameOver {
		var textStr string
		var col color.Color
		if state.HasWon {
			textStr = winText
			col = winColor
		} else {
			textStr = gameOverText
			col = gameOverColor
		}
		bounds, _ := font.BoundString(r.font, textStr)
		textWidth := (bounds.Max.X - bounds.Min.X).Ceil()
		text.Draw(screen, textStr, r.font, width/2-textWidth/2, headerHeight/2+6, col)
	}
}

func (r *EbitenRenderer) screenToGamePos(x, y int) (game.Position, bool) {
	if r.game == nil {
		return game.Position{}, false
	}
	invTransform := r.transform
	invTransform.Invert()
	gx, gy := invTransform.Apply(float64(x), float64(y))
	pos := game.Position{
		int(gx) / r.cellSize,
		int(gy) / r.cellSize,
	}
	if pos.X >= 0 && pos.X < r.game.Width() && pos.Y >= 0 && pos.Y < r.game.Height() {
		return pos, true
	}
	return game.Position{}, false
}

func (r *EbitenRenderer) mouseButtonClicked() (game.Position, bool) {
	x, y := ebiten.CursorPosition()
	return r.screenToGamePos(x, y)
}

func (r *EbitenRenderer) drawCell(screen *ebiten.Image, pos game.Position, yOffset int) {
	if r.game == nil {
		return
	}
	scale := float32(1.0)
	state := r.game.GetCellState(pos)

	if anim, animating := r.animations[pos]; animating {
		elapsed := time.Since(anim.StartTime)
		if elapsed < 0 {
			scale = 0.5
		} else {
			progress := float64(elapsed) / float64(anim.Duration)
			if progress > 1.0 {
				progress = 1.0
			}
			scale = float32(0.5 + 0.5*progress)
		}
	}

	xPos := float32(pos.X * r.cellSize)
	yPos := float32(pos.Y*r.cellSize) + float32(yOffset)
	size := float32(r.cellSize)

	centerX := xPos + size/2
	centerY := yPos + size/2

	scaledSize := size * scale
	drawX := centerX - scaledSize/2
	drawY := centerY - scaledSize/2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(drawX), float64(drawY))

	switch state {
	case game.StateHidden:
		screen.DrawImage(r.sprites.hidden, op)
	case game.StateRevealed:
		if _, animating := r.animations[pos]; animating {
			screen.DrawImage(r.sprites.hidden, op)
		} else {
			screen.DrawImage(r.sprites.revealed, op)
			if r.game.GetCellContent(pos) == game.ContentMine {
				screen.DrawImage(r.sprites.mine, op)
			} else {
				mineCount := r.game.GetAdjacentMines(pos)
				if mineCount > 0 {
					numberStr := strconv.Itoa(mineCount)
					bounds, _ := font.BoundString(r.font, numberStr)
					textWidth := (bounds.Max.X - bounds.Min.X).Ceil()
					textHeight := (bounds.Max.Y - bounds.Min.Y).Ceil()

					textX := int(xPos) + (r.cellSize-textWidth)/2
					textY := int(yPos) + (r.cellSize+textHeight)/2
					text.Draw(screen, numberStr, r.font, textX, textY, numberColours[mineCount])
				}
			}
		}
	case game.StateFlagged:
		screen.DrawImage(r.sprites.hidden, op)
		screen.DrawImage(r.sprites.flag, op)
	}
}

var numberColours = []color.Color{
	color.RGBA{128, 128, 128, 255},
	color.RGBA{255, 255, 255, 255},
	color.RGBA{118, 253, 75, 255},
	color.RGBA{255, 255, 85, 255},
	color.RGBA{255, 0, 0, 255},
	color.RGBA{128, 0, 0, 255},
	color.RGBA{0, 128, 128, 255},
	color.RGBA{0, 0, 0, 255},
	color.RGBA{128, 128, 128, 255},
}
