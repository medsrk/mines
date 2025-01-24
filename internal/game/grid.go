package game

import (
	"math/rand"
)

type Grid struct {
	width, height int
	mines         map[Position]struct{}
	states        [][]CellState
}

func NewGrid(width, height int) *Grid {
	g := &Grid{
		width:  width,
		height: height,
		mines:  make(map[Position]struct{}),
		states: make([][]CellState, width),
	}

	for i := 0; i < width; i++ {
		g.states[i] = make([]CellState, height)
	}

	return g
}

func (g *Grid) PlaceMines(count int, exclude Position) {
	available := make([]Position, 0, g.width*g.height)

	for x := 0; x < g.width; x++ {
		for y := 0; y < g.height; y++ {
			pos := Position{x, y}
			if pos != exclude {
				available = append(available, pos)
			}
		}
	}

	rand.Shuffle(len(available), func(i, j int) {
		available[i], available[j] = available[j], available[i]
	})

	g.mines = make(map[Position]struct{}, count)
	for _, pos := range available[:count] {
		g.mines[pos] = struct{}{}
	}
}

func (g *Grid) GetAdjacentMines(p Position) int {
	count := 0
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}

			nx, ny := p.X+dx, p.Y+dy
			if nx < 0 || nx >= g.width || ny < 0 || ny >= g.height {
				continue
			}

			if _, isMine := g.mines[Position{nx, ny}]; isMine {
				count++
			}
		}
	}
	return count
}

func (g *Grid) Reveal(p Position) {
	if _, isMine := g.mines[p]; isMine {
		g.states[p.X][p.Y] = StateRevealed
		return
	}

	if g.states[p.X][p.Y] != StateHidden {
		return
	}

	g.states[p.X][p.Y] = StateRevealed

	if g.GetAdjacentMines(p) == 0 {
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				nx, ny := p.X+dx, p.Y+dy
				if nx >= 0 && nx < g.width && ny >= 0 && ny < g.height {
					g.Reveal(Position{nx, ny})
				}
			}
		}
	}
}
