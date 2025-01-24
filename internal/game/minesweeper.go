package game

type Minesweeper struct {
	grid      *Grid
	gameState GameState
}

func NewMinesweeper(width, height int) *Minesweeper {
	return &Minesweeper{
		grid: NewGrid(width, height),
	}
}

func (ms *Minesweeper) HandleLeftClick(p Position) {
	if len(ms.grid.mines) == 0 {
		ms.grid.PlaceMines(20, p)
	}

	ms.grid.Reveal(p)

	if _, isMine := ms.grid.mines[p]; isMine {
		ms.gameState.IsGameOver = true
	}
}

func (ms *Minesweeper) HandleRightClick(p Position) {
	if !ms.gameState.IsGameOver {
		switch ms.grid.states[p.X][p.Y] {
		case StateHidden:
			ms.grid.states[p.X][p.Y] = StateFlagged
		case StateFlagged:
			ms.grid.states[p.X][p.Y] = StateQuestion
		default:
			ms.grid.states[p.X][p.Y] = StateHidden
		}
	}
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
