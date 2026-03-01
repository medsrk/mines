package game

import "time"

type Position struct {
	X, Y int
}

type CellState int

const (
	StateHidden CellState = iota
	StateRevealed
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

type EventType int

const (
	EventReveal EventType = iota
	EventExplosion
	EventFlagged
	EventUnflagged
	EventWon
)

type Event struct {
	Type      EventType
	Positions []Position
}

type Game interface {
	Width() int
	Height() int
	GetCellState(Position) CellState
	GetCellContent(Position) CellContent
	GetAdjacentMines(Position) int
	GameState() GameState

	HandleLeftClick(Position) []Event
	HandleRightClick(Position) []Event

	Update() []Event
	MineCount() int
	FlagCount() int
}
