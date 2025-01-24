package game

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

type CellContent int

const (
	ContentEmpty CellContent = iota
	ContentMine
)

type GameState struct {
	IsGameOver bool
	HasWon     bool
}

type Game interface {
	Width() int
	Height() int
	GetCellState(Position) CellState
	GetCellContent(Position) CellContent
	GetAdjacentMines(Position) int
	GameState() GameState

	HandleLeftClick(Position)
	HandleRightClick(Position)
}
