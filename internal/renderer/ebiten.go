package renderer

import (
	"bytes"
	"fmt"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image/color"
	"io"
	"log"
	"mines/internal/game"
	"os"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
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

var (
	customWidth  int = 10
	customHeight int = 10
	customMines  int = 10
)

type EbitenRenderer struct {
	game      game.Game
	cellSize  int
	font      font.Face
	sprites   *Sprites
	transform ebiten.GeoM
	menuState MenuState
	menuItems []string

	audioCtx           *audio.Context
	tileRevealSound    *audio.Player
	initialRevealSound *audio.Player
	flagSound          *audio.Player
	unflagSound        *audio.Player
	explodeSound       *audio.Player
	winSound           *audio.Player

	prevHasWon bool
}

func loadSound(audioCtx *audio.Context, path string) *audio.Player {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Warning: loadSound: failed to open sound file %s: %v", path, err)
		return nil
	}
	defer file.Close()

	decodedStream, err := vorbis.DecodeWithSampleRate(audioCtx.SampleRate(), file)
	if err != nil {
		log.Printf("Warning: loadSound: failed to decode OGG sound file %s: %v", path, err)
		return nil
	}

	var pcmBuffer bytes.Buffer
	if _, err := io.Copy(&pcmBuffer, decodedStream); err != nil {
		log.Printf("Warning: loadSound: failed to buffer decoded OGG stream for %s: %v", path, err)
		return nil
	}

	pcmDataReader := bytes.NewReader(pcmBuffer.Bytes())
	player, err := audioCtx.NewPlayer(pcmDataReader)
	if err != nil {
		log.Printf("Warning: loadSound: failed to create audio player for %s from buffered data: %v", path, err)
		return nil
	}
	return player
}

func NewEbitenRenderer(g game.Game, cellSize int) *EbitenRenderer {
	sprites, err := LoadSprites()
	if err != nil {
		panic(fmt.Sprintf("failed to load sprites: %v", err))
	}

	var transform ebiten.GeoM
	transform.Translate(0, headerHeight)

	menuItems := make([]string, len(difficulties))
	for i, d := range difficulties {
		menuItems[i] = fmt.Sprintf("%s (%dx%d, %d mines)", d.Name, d.Width, d.Height, d.MineCount)
	}

	audioCtx := audio.NewContext(44100)

	tileRevealSound := loadSound(audioCtx, "assets/audio/reveal.ogg")
	initialRevealSound := loadSound(audioCtx, "assets/audio/click.ogg")
	flagSound := loadSound(audioCtx, "assets/audio/flag.ogg")
	unflagSound := loadSound(audioCtx, "assets/audio/unflag.ogg")
	explodeSound := loadSound(audioCtx, "assets/audio/explode.ogg")
	winSound := loadSound(audioCtx, "assets/audio/win.ogg")

	return &EbitenRenderer{
		game:               g,
		cellSize:           cellSize,
		font:               basicfont.Face7x13,
		sprites:            sprites,
		transform:          transform,
		menuState:          MenuStateMain,
		menuItems:          menuItems,
		audioCtx:           audioCtx,
		tileRevealSound:    tileRevealSound,
		initialRevealSound: initialRevealSound,
		flagSound:          flagSound,
		unflagSound:        unflagSound,
		explodeSound:       explodeSound,
		winSound:           winSound,
		prevHasWon:         false,
	}
}

func (r *EbitenRenderer) Update() error {
	if r.menuState == MenuStateMain {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			x, y := ebiten.CursorPosition()

			// Handle preset difficulty selection
			for i, item := range r.menuItems {
				bounds, _ := font.BoundString(r.font, item)
				width := (bounds.Max.X - bounds.Min.X).Ceil()
				itemX := screenWidth/2 - width/2
				itemY := 200 + i*50
				if x >= itemX && x <= itemX+width && y >= itemY-20 && y <= itemY+20 {
					d := difficulties[i]
					r.game = game.NewMinesweeper(d.Width, d.Height, d.MineCount)
					r.menuState = MenuStatePlaying
					r.prevHasWon = false
					ebiten.SetWindowSize(d.Width*r.cellSize, d.Height*r.cellSize+headerHeight)
					winWidth, winHeight := ebiten.Monitor().Size()
					ebiten.SetWindowPosition((winWidth-d.Width*r.cellSize)/2, (winHeight-d.Height*r.cellSize-headerHeight)/2)
					return nil
				}
			}

			// Handle custom settings buttons
			baseY := 200 + len(difficulties)*50 + 30
			rows := []struct {
				valuePtr *int
				minValue int
				maxValue int
			}{
				{valuePtr: &customWidth, minValue: 5, maxValue: 50},
				{valuePtr: &customHeight, minValue: 5, maxValue: 50},
				{valuePtr: &customMines, minValue: 1, maxValue: customWidth*customHeight - 1},
			}

			for i := range rows {
				field := []struct {
					label string
					value int
				}{
					{"Width:", customWidth}, {"Height:", customHeight}, {"Mines:", customMines},
				}[i]
				yPos := baseY + i*menuRowSpacing
				valueStr := fmt.Sprintf("%d", *rows[i].valuePtr)
				bounds, _ := font.BoundString(r.font, field.label)
				labelWidth := (bounds.Max.X - bounds.Min.X).Ceil()
				bounds, _ = font.BoundString(r.font, valueStr)
				valueWidth := (bounds.Max.X - bounds.Min.X).Ceil()
				totalWidth := labelWidth + menuButtonPadding + valueWidth + menuButtonPadding + menuButtonSize*2 + menuButtonPadding
				labelX := screenWidth/2 - totalWidth/2
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
			startY := baseY + 3*menuRowSpacing
			startText := "Start Custom Game"
			bounds, _ := font.BoundString(r.font, startText)
			startWidth := (bounds.Max.X - bounds.Min.X).Ceil()
			startX := screenWidth/2 - startWidth/2
			if x >= startX-10 && x <= startX+startWidth+10 && y >= startY && y <= startY+30 {
				r.game = game.NewMinesweeper(customWidth, customHeight, customMines)
				r.menuState = MenuStatePlaying
				r.prevHasWon = false
				ebiten.SetWindowSize(customWidth*r.cellSize, customHeight*r.cellSize+headerHeight)
				winWidth, winHeight := ebiten.Monitor().Size()
				ebiten.SetWindowPosition((winWidth-customWidth*r.cellSize)/2, (winHeight-customHeight*r.cellSize-headerHeight)/2)
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
			pos, clicked := r.mouseButtonClicked(ebiten.MouseButtonLeft)
			if clicked {
				_, clickResult := r.game.HandleLeftClick(pos)

				switch clickResult {
				case game.ResultExplosion:
					if r.explodeSound != nil {
						r.explodeSound.Rewind()
						r.explodeSound.Play()
					}
				case game.ResultReveal:
					if r.initialRevealSound != nil {
						r.initialRevealSound.Rewind()
						r.initialRevealSound.Play()
					}
				}
			}
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			pos, clicked := r.mouseButtonClicked(ebiten.MouseButtonRight)
			if clicked {
				r.game.HandleRightClick(pos)
				if r.game.GetCellState(pos) == game.StateFlagged {
					if r.flagSound != nil {
						r.flagSound.Rewind()
						r.flagSound.Play()
					}
				}
				if r.game.GetCellState(pos) == game.StateHidden {
					if r.unflagSound != nil {
						r.unflagSound.Rewind()
						r.unflagSound.Play()
					}
				}
			}
		}
	}

	// --- Game State Update and Per-Tile Animation Sounds ---
	justFullyRevealed := r.game.Update()

	if r.tileRevealSound != nil {
		for _, _ = range justFullyRevealed {
			r.tileRevealSound.Rewind()
			r.tileRevealSound.Play()
		}
	}

	// --- Win Condition Sound ---
	currentGameState := r.game.GameState()
	if currentGameState.HasWon && !r.prevHasWon {
		if r.winSound != nil {
			r.winSound.Rewind()
			r.winSound.Play()
		}
	}
	r.prevHasWon = currentGameState.HasWon

	// --- Reset Game ---
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		r.menuState = MenuStateMain
		ebiten.SetWindowSize(1280, 720)
		winWidth, winHeight := ebiten.Monitor().Size()
		ebiten.SetWindowPosition((winWidth-1280)/2, (winHeight-720)/2)
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
	// Draw menu title

	bounds, _ := font.BoundString(r.font, menuText)
	text.Draw(screen, menuText, r.font, screenWidth/2-(bounds.Max.X-bounds.Min.X).Ceil()/2, 100, color.White)

	// Draw preset difficulty options
	for i, item := range r.menuItems {
		bounds, _ := font.BoundString(r.font, item)
		text.Draw(screen, item, r.font, screenWidth/2-(bounds.Max.X-bounds.Min.X).Ceil()/2, 200+i*50, color.White)
	}

	// Calculate base Y dynamically
	baseY := 200 + len(difficulties)*50 + 30
	for i, field := range []struct {
		label string
		value int
	}{
		{"Width:", customWidth}, {"Height:", customHeight}, {"Mines:", customMines},
	} {
		y := baseY + i*menuRowSpacing
		valueStr := fmt.Sprintf("%d", field.value)
		bounds, _ := font.BoundString(r.font, field.label)
		labelWidth := (bounds.Max.X - bounds.Min.X).Ceil()
		bounds, _ = font.BoundString(r.font, valueStr)
		valueWidth := (bounds.Max.X - bounds.Min.X).Ceil()
		totalWidth := labelWidth + menuButtonPadding + valueWidth + menuButtonPadding + menuButtonSize*2 + menuButtonPadding
		labelX := screenWidth/2 - totalWidth/2

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
	startY := baseY + 3*menuRowSpacing
	startText := "Start Custom Game"
	bounds, _ = font.BoundString(r.font, startText)
	startWidth := (bounds.Max.X - bounds.Min.X).Ceil()
	startX := screenWidth/2 - startWidth/2
	vector.DrawFilledRect(screen, float32(startX)-10, float32(startY), float32(startWidth)+20, 30, color.RGBA{50, 150, 50, 255}, false)
	text.Draw(screen, startText, r.font, startX, startY+20, color.White)
}

func (r *EbitenRenderer) Layout(w, h int) (int, int) {
	if r.menuState == MenuStateMain || r.game == nil {
		return 1280, 720
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

func (r *EbitenRenderer) mouseButtonClicked(button ebiten.MouseButton) (game.Position, bool) {
	if r.game == nil {
		return game.Position{}, false
	}
	x, y := ebiten.CursorPosition()
	invTransform := r.transform
	invTransform.Invert()
	gx, gy := invTransform.Apply(float64(x), float64(y))

	pos := game.Position{
		X: int(gx) / r.cellSize,
		Y: int(gy) / r.cellSize,
	}

	if pos.X >= 0 && pos.X < r.game.Width() && pos.Y >= 0 && pos.Y < r.game.Height() {
		return pos, true
	}
	return game.Position{}, false
}

func (r *EbitenRenderer) drawCell(screen *ebiten.Image, pos game.Position, yOffset int) {
	if r.game == nil {
		return
	}
	scale := float32(1.0)
	state := r.game.GetCellState(pos)

	if state == game.StateRevealing {
		if anim, exists := r.game.Grid().GetAnimation(pos); exists {
			elapsed := time.Since(anim.StartTime)
			if elapsed >= 0 {
				progress := float64(elapsed) / float64(anim.Duration)
				if progress < 0 {
					progress = 0
				}
				if progress > 1.0 {
					progress = 1.0
				}
				scale = float32(0.5 + 0.5*progress)
			} else {
				scale = 0.5
			}
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
				textWidth := (bounds.Max.X - bounds.Min.X).Ceil()
				textHeight := (bounds.Max.Y - bounds.Min.Y).Ceil()

				textX := int(xPos) + (r.cellSize-textWidth)/2
				textY := int(yPos) + (r.cellSize+textHeight)/2
				text.Draw(screen, numberStr, r.font, textX, textY, numberColours[mineCount])
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
