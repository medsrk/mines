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

type ClickResultType int

const (
	ResultNoOp   ClickResultType = iota
	ResultReveal                 // Sound for initial reveal action
	ResultExplosion
)

type Game interface {
	Width() int
	Height() int
	GetCellState(Position) CellState
	GetCellContent(Position) CellContent
	GetAdjacentMines(Position) int
	GameState() GameState

	HandleLeftClick(Position) ([]Position, ClickResultType)
	HandleRightClick(Position) bool

	Update() []Position
	Grid() *Grid
	MineCount() int
	FlagCount() int
}

type Animation struct {
	StartTime time.Time
	Duration  time.Duration
}
