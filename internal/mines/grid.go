package mines

import (
	"iter"
	"math/rand"
	"slices"
)

type Position struct {
	X, Y int
}

type CellState int

const (
	StateHidden CellState = iota
	StateRevealed
	StateFlagged
	StateQuestion
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
	// Create slice of positions using the iterator
	available := slices.Collect(g.Positions())

	// Remove excluded position
	available = slices.DeleteFunc(available, func(p Position) bool {
		return p == exclude
	})

	// Shuffle and take first count positions
	rand.Shuffle(len(available), func(i, j int) {
		available[i], available[j] = available[j], available[i]
	})

	g.mines = make(map[Position]struct{}, count)
	for _, pos := range available[:count] {
		g.mines[pos] = struct{}{}
	}
}

// Positions returns an iterator over all positions in the grid
func (g *Grid) Positions() iter.Seq[Position] {
	return func(yield func(Position) bool) {
		for x := 0; x < g.width; x++ {
			for y := 0; y < g.height; y++ {
				if !yield(Position{x, y}) {
					return
				}
			}
		}
	}
}

// Neighbors returns an iterator over all adjacent positions
func (g *Grid) Neighbors(p Position) iter.Seq[Position] {
	return func(yield func(Position) bool) {
		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				if dx == 0 && dy == 0 {
					continue
				}

				x, y := p.X+dx, p.Y+dy
				if x < 0 || x >= g.width || y < 0 || y >= g.height {
					continue
				}

				if !yield(Position{x, y}) {
					return
				}
			}
		}
	}
}

func (g *Grid) Neighbors2(p Position) []Position {
	neighbors := make([]Position, 0, 8)
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}

			x, y := p.X+dx, p.Y+dy
			if x < 0 || x >= g.width || y < 0 || y >= g.height {
				continue
			}

			neighbors = append(neighbors, Position{x, y})
		}
	}
	return neighbors
}

func (g *Grid) GetAdjacentMines(p Position) int {
	count := 0
	for _, n := range g.Neighbors2(p) {
		if _, ok := g.mines[n]; ok {
			count++
		}
	}
	return count
}

func (g *Grid) Reveal(p Position) {

}
