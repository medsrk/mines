package mines

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"image/color"
)

var (
	LightGrey = color.RGBA{200, 200, 200, 255}
	DarkGrey  = color.RGBA{100, 100, 100, 255}
	Yellow    = color.RGBA{255, 255, 0, 255}
	Red       = color.RGBA{255, 0, 0, 255}
	Green     = color.RGBA{0, 255, 0, 255}
	Blue      = color.RGBA{0, 0, 255, 255}
)

type Game struct {
	Grid *Grid
}

func NewGame() *Game {
	grid := NewGrid(10, 10)
	return &Game{
		Grid: grid,
	}
}

func (g *Game) Update() error {
	x, y := ebiten.CursorPosition()

	p := Position{
		X: x / 50,
		Y: y / 50,
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if len(g.Grid.mines) == 0 {
			g.Grid.PlaceMines(10, p)
		}
		n := g.Grid.GetAdjacentMines(p)
		fmt.Println(g.Grid)
		fmt.Println(n)

	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(LightGrey)
	g.Grid.Positions()(func(p Position) bool {
		x, y := p.X*50, p.Y*50
		vector.DrawFilledRect(screen, float32(x), float32(y), 50, 50, DarkGrey, true)
		vector.StrokeRect(screen, float32(x), float32(y), 50, 50, 1, Yellow, true)
		if _, ok := g.Grid.mines[p]; ok {
			vector.DrawFilledCircle(screen, float32(x+25), float32(y+25), 20, Red, true)
		}
		return true
	})
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 800, 600
}
