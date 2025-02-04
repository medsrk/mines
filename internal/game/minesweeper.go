package game

type Minesweeper struct {
	grid      *Grid
	gameState GameState
	mineCount int
}

func NewMinesweeper(width, height, mineCount int) *Minesweeper {
	return &Minesweeper{
		grid:      NewGrid(width, height),
		mineCount: mineCount,
	}
}

func (ms *Minesweeper) HandleLeftClick(p Position) []Position {
	if len(ms.grid.mines) == 0 {
		ms.grid.PlaceMines(ms.mineCount, p)
	}

	revealed := ms.grid.Reveal(p, p)

	if _, isMine := ms.grid.mines[p]; isMine {
		ms.gameState.IsGameOver = true
	}

	return revealed
}

func (ms *Minesweeper) HandleRightClick(p Position) {
	if !ms.gameState.IsGameOver {
		posState := ms.grid.states[p.X][p.Y]
		if posState == StateRevealed {
			return
		}
		switch posState {
		case StateHidden:
			ms.grid.states[p.X][p.Y] = StateFlagged
		default:
			ms.grid.states[p.X][p.Y] = StateHidden
		}
	}
}

func (ms *Minesweeper) Update() {
	ms.grid.Update()
}

func (ms *Minesweeper) Width() int { return ms.grid.width }

func (ms *Minesweeper) Height() int { return ms.grid.height }

func (ms *Minesweeper) GetCellState(p Position) CellState { return ms.grid.states[p.X][p.Y] }

func (ms *Minesweeper) GetCellContent(p Position) CellContent {
	if _, isMine := ms.grid.mines[p]; isMine {
		return ContentMine
	}
	return ContentEmpty
}

func (ms *Minesweeper) GetAdjacentMines(p Position) int { return ms.grid.GetAdjacentMines(p) }

func (ms *Minesweeper) GameState() GameState { return ms.gameState }

func (ms *Minesweeper) Grid() *Grid { return ms.grid }
