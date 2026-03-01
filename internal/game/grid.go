package game

import (
	"iter"
	"math/rand"
)

type Grid struct {
	width, height int
	mines         map[Position]struct{}
	flags         map[Position]struct{}
	states        [][]CellState
}

func NewGrid(width, height int) *Grid {
	g := &Grid{
		width:  width,
		height: height,
		mines:  make(map[Position]struct{}),
		flags:  make(map[Position]struct{}),
		states: make([][]CellState, width),
	}

	for i := 0; i < width; i++ {
		g.states[i] = make([]CellState, height)
	}

	return g
}

func (g *Grid) PlaceMines(count int, exclude Position) {
	excluded := make(map[Position]struct{})
	excluded[exclude] = struct{}{}

	for neighbor := range exclude.Neighbors(g.width, g.height) {
		excluded[neighbor] = struct{}{}
	}

	available := make([]Position, 0, g.width*g.height-len(excluded))
	for x := 0; x < g.width; x++ {
		for y := 0; y < g.height; y++ {
			pos := Position{x, y}
			if _, isExcluded := excluded[pos]; !isExcluded {
				available = append(available, pos)
			}
		}
	}

	if count > len(available) {
		count = len(available)
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
	for neighbor := range p.Neighbors(g.width, g.height) {
		if _, isMine := g.mines[neighbor]; isMine {
			count++
		}
	}
	return count
}

func (g *Grid) Reveal(p Position) []Position {
	var revealed []Position

	// If already revealed and has adjacent mines, try to chord
	if g.states[p.X][p.Y] == StateRevealed {
		adjacentMines := g.GetAdjacentMines(p)
		if adjacentMines > 0 {
			flagCount := 0
			for neighbor := range p.Neighbors(g.width, g.height) {
				if g.states[neighbor.X][neighbor.Y] == StateFlagged {
					flagCount++
				}
			}

			// If flag count matches, reveal non-flagged neighbors
			if flagCount == adjacentMines {
				for neighbor := range p.Neighbors(g.width, g.height) {
					if g.states[neighbor.X][neighbor.Y] != StateFlagged && g.states[neighbor.X][neighbor.Y] != StateRevealed {
						revealed = append(revealed, g.Reveal(neighbor)...)
					}
				}
			}
		}
		return revealed
	}

	if g.states[p.X][p.Y] != StateHidden {
		return revealed
	}

	g.states[p.X][p.Y] = StateRevealed

	revealed = append(revealed, p)

	// If empty cell, cascade to neighbors
	if _, isMine := g.mines[p]; !isMine && g.GetAdjacentMines(p) == 0 {
		for neighbor := range p.Neighbors(g.width, g.height) {
			if _, isMine := g.mines[neighbor]; !isMine {
				revealed = append(revealed, g.Reveal(neighbor)...)
			}
		}
	}

	return revealed
}

func (g *Grid) ToggleFlag(p Position) {
	if g.states[p.X][p.Y] == StateHidden {
		g.flags[p] = struct{}{}
		g.states[p.X][p.Y] = StateFlagged
	} else if g.states[p.X][p.Y] == StateFlagged {
		delete(g.flags, p)
		g.states[p.X][p.Y] = StateHidden
	}
}

func (p Position) Neighbors(width, height int) iter.Seq[Position] {
	return func(yield func(Position) bool) {
		deltas := []struct{ dx, dy int }{
			{-1, -1}, {-1, 0}, {-1, 1},
			{0, -1}, {0, 1},
			{1, -1}, {1, 0}, {1, 1},
		}

		for _, d := range deltas {
			nx, ny := p.X+d.dx, p.Y+d.dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				if !yield(Position{nx, ny}) {
					return
				}
			}
		}
	}
}

func (g *Grid) CheckHasWon() bool {
	for x := 0; x < g.width; x++ {
		for y := 0; y < g.height; y++ {
			pos := Position{x, y}
			if _, isMine := g.mines[pos]; !isMine && g.states[x][y] != StateRevealed {
				return false
			}
		}
	}
	return true
}
