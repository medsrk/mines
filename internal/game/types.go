package game

import "time"

type Position struct {
	X, Y int
}

type CellState int

const (
	StateHidden CellState = iota
	StateRevealed
	StateRevealing
	StateFlagged
)

type CellContent int

const (
	ContentEmpty CellContent = iota
	ContentMine
)

type GameState struct {
	IsGameOver  bool
	HasWon      bool
	StartTime   time.Time
	ElapsedTime time.Duration
}

type Game interface {
	Width() int
	Height() int
	GetCellState(Position) CellState
	GetCellContent(Position) CellContent
	GetAdjacentMines(Position) int
	GameState() GameState

	HandleLeftClick(Position) []Position
	HandleRightClick(Position)

	Update()
	Grid() *Grid
	MineCount() int
	FlagCount() int
}

type AnimationType int

const (
	AnimationReveal AnimationType = iota
	AnimationFlag
)

type Animation struct {
	Type      AnimationType
	StartTime time.Time
	Duration  time.Duration
}
