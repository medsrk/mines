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

func (ms *Minesweeper) HandleLeftClick(p Position) ([]Position, ClickResultType) {
	if ms.gameState.IsGameOver || ms.gameState.HasWon {
		return nil, ResultNoOp
	}
	if ms.grid.states[p.X][p.Y] == StateFlagged {
		return nil, ResultNoOp
	}

	if len(ms.grid.mines) == 0 {
		ms.grid.PlaceMines(ms.mineCount, p)
	}

	revealed := ms.grid.Reveal(p, p)

	hitMine := false
	if len(revealed) > 0 {
		for _, rPos := range revealed {
			if _, isMine := ms.grid.mines[rPos]; isMine {
				hitMine = true
				break
			}
		}
	}

	if hitMine {
		ms.gameState.IsGameOver = true
		return revealed, ResultExplosion
	}

	if len(revealed) > 0 {
		return revealed, ResultReveal
	}

	return nil, ResultNoOp
}

func (ms *Minesweeper) HandleRightClick(p Position) bool {
	if ms.gameState.IsGameOver || ms.gameState.HasWon {
		return false
	}

	posState := ms.grid.states[p.X][p.Y]
	if posState == StateRevealed {
		return false
	}

	if posState == StateHidden {
		ms.grid.ToggleFlag(p)
		return true
	} else if posState == StateFlagged {
		ms.grid.ToggleFlag(p)
		return true
	}

	return false
}

func (ms *Minesweeper) Update() []Position {
	justFullyRevealed := ms.grid.Update()

	if !ms.gameState.IsGameOver && !ms.gameState.StartTime.IsZero() {
		ms.gameState.ElapsedTime = time.Since(ms.gameState.StartTime)
	}

	if !ms.gameState.IsGameOver {
		if ms.grid.CheckHasWon() {
			ms.gameState.HasWon = true
			ms.gameState.IsGameOver = true
		}
	}

	if ms.gameState.IsGameOver && !ms.gameState.HasWon {
		for minePos := range ms.grid.mines {
			if ms.grid.states[minePos.X][minePos.Y] == StateHidden {
				ms.grid.states[minePos.X][minePos.Y] = StateRevealed
			}
		}
	}

	return justFullyRevealed
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
func (ms *Minesweeper) Grid() *Grid                     { return ms.grid }
func (ms *Minesweeper) MineCount() int                  { return ms.mineCount }
func (ms *Minesweeper) FlagCount() int                  { return len(ms.grid.flags) }
