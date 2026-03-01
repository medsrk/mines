package game

import "time"

type Minesweeper struct {
	grid      *Grid
	gameState GameState
	mineCount int
}

func NewMinesweeper(width, height, mineCount int) *Minesweeper {
	return &Minesweeper{
		grid:      NewGrid(width, height),
		mineCount: mineCount,
		gameState: GameState{
			StartTime: time.Now(),
		},
	}
}

func (ms *Minesweeper) HandleLeftClick(p Position) []Event {
	if ms.gameState.IsGameOver || ms.gameState.HasWon {
		return nil
	}
	if ms.grid.states[p.X][p.Y] == StateFlagged {
		return nil
	}

	if len(ms.grid.mines) == 0 {
		ms.grid.PlaceMines(ms.mineCount, p)
	}

	revealed := ms.grid.Reveal(p)
	if len(revealed) == 0 {
		return nil
	}

	for _, rPos := range revealed {
		if _, isMine := ms.grid.mines[rPos]; isMine {
			ms.gameState.IsGameOver = true
			return []Event{{EventExplosion, []Position{p}}}
		}
	}

	return []Event{{EventReveal, revealed}}
}

func (ms *Minesweeper) HandleRightClick(p Position) []Event {
	if ms.gameState.IsGameOver || ms.gameState.HasWon {
		return nil
	}

	posState := ms.grid.states[p.X][p.Y]
	switch posState {
	case StateHidden:
		ms.grid.ToggleFlag(p)
		return []Event{{EventFlagged, []Position{p}}}
	case StateFlagged:
		ms.grid.ToggleFlag(p)
		return []Event{{EventUnflagged, []Position{p}}}
	}

	return nil
}

func (ms *Minesweeper) Update() []Event {
	var events []Event

	if !ms.gameState.IsGameOver && !ms.gameState.StartTime.IsZero() {
		ms.gameState.ElapsedTime = time.Since(ms.gameState.StartTime)
	}

	if !ms.gameState.IsGameOver {
		if ms.grid.CheckHasWon() {
			ms.gameState.HasWon = true
			ms.gameState.IsGameOver = true
			events = append(events, Event{Type: EventWon})
		}
	}

	if ms.gameState.IsGameOver && !ms.gameState.HasWon {
		for minePos := range ms.grid.mines {
			if ms.grid.states[minePos.X][minePos.Y] == StateHidden {
				ms.grid.states[minePos.X][minePos.Y] = StateRevealed
			}
		}
	}

	return events
}

func (ms *Minesweeper) Width() int                        { return ms.grid.width }
func (ms *Minesweeper) Height() int                       { return ms.grid.height }
func (ms *Minesweeper) GetCellState(p Position) CellState { return ms.grid.states[p.X][p.Y] }
func (ms *Minesweeper) GetCellContent(p Position) CellContent {
	if _, isMine := ms.grid.mines[p]; isMine {
		return ContentMine
	}
	return ContentEmpty
}
func (ms *Minesweeper) GetAdjacentMines(p Position) int { return ms.grid.GetAdjacentMines(p) }
func (ms *Minesweeper) GameState() GameState            { return ms.gameState }
func (ms *Minesweeper) MineCount() int                  { return ms.mineCount }
func (ms *Minesweeper) FlagCount() int                  { return len(ms.grid.flags) }
