package game

import (
	"fmt"
	"strings"
)

type CellState int

const (
	Empty CellState = iota
	X
	O
)

func (c CellState) String() string {
	switch c {
	case X:
		return "X"
	case O:
		return "O"
	default:
		return " "
	}
}

type BoardState int

const (
	Undecided BoardState = iota
	XWins
	OWins
	Draw
)

func (b BoardState) String() string {
	switch b {
	case XWins:
		return "X"
	case OWins:
		return "O"
	case Draw:
		return "D"
	default:
		return " "
	}
}

type SmallBoard struct {
	Cells [9]CellState
	State BoardState
}

func NewSmallBoard() *SmallBoard {
	return &SmallBoard{
		State: Undecided,
	}
}

func (sb *SmallBoard) IsValidMove(position int) bool {
	return position >= 0 && position < 9 && sb.Cells[position] == Empty && sb.State == Undecided
}

func (sb *SmallBoard) MakeMove(position int, player CellState) bool {
	if !sb.IsValidMove(position) {
		return false
	}
	sb.Cells[position] = player
	sb.updateState()
	return true
}

func (sb *SmallBoard) updateState() {
	for i := 0; i < 3; i++ {
		if sb.Cells[i*3] != Empty &&
			sb.Cells[i*3] == sb.Cells[i*3+1] &&
			sb.Cells[i*3+1] == sb.Cells[i*3+2] {
			if sb.Cells[i*3] == X {
				sb.State = XWins
			} else {
				sb.State = OWins
			}
			return
		}
	}
	for i := 0; i < 3; i++ {
		if sb.Cells[i] != Empty &&
			sb.Cells[i] == sb.Cells[i+3] &&
			sb.Cells[i+3] == sb.Cells[i+6] {
			if sb.Cells[i] == X {
				sb.State = XWins
			} else {
				sb.State = OWins
			}
			return
		}
	}
	if sb.Cells[0] != Empty && sb.Cells[0] == sb.Cells[4] && sb.Cells[4] == sb.Cells[8] {
		if sb.Cells[0] == X {
			sb.State = XWins
		} else {
			sb.State = OWins
		}
		return
	}
	if sb.Cells[2] != Empty && sb.Cells[2] == sb.Cells[4] && sb.Cells[4] == sb.Cells[6] {
		if sb.Cells[2] == X {
			sb.State = XWins
		} else {
			sb.State = OWins
		}
		return
	}
	full := true
	for _, cell := range sb.Cells {
		if cell == Empty {
			full = false
			break
		}
	}
	if full {
		sb.State = Draw
	}
}

func (sb *SmallBoard) IsFull() bool {
	for _, cell := range sb.Cells {
		if cell == Empty {
			return false
		}
	}
	return true
}

type UltimateBoard struct {
	Boards      [9]*SmallBoard
	State       BoardState
	ActiveBoard int
	CurrentTurn CellState
}

func NewUltimateBoard() *UltimateBoard {
	ub := &UltimateBoard{
		State:       Undecided,
		ActiveBoard: -1,
		CurrentTurn: X,
	}
	for i := 0; i < 9; i++ {
		ub.Boards[i] = NewSmallBoard()
	}
	return ub
}

func (ub *UltimateBoard) IsValidMove(boardIndex, position int) bool {
	if boardIndex < 0 || boardIndex >= 9 {
		return false
	}
	if ub.ActiveBoard != -1 && ub.ActiveBoard != boardIndex {
		return false
	}
	if ub.Boards[boardIndex].State != Undecided {
		return false
	}
	return ub.Boards[boardIndex].IsValidMove(position)
}

func (ub *UltimateBoard) MakeMove(boardIndex, position int) error {
	if !ub.IsValidMove(boardIndex, position) {
		return fmt.Errorf("invalid move: board %d, position %d", boardIndex, position)
	}
	if !ub.Boards[boardIndex].MakeMove(position, ub.CurrentTurn) {
		return fmt.Errorf("failed to make move on board %d, position %d", boardIndex, position)
	}
	if ub.Boards[position].State == Undecided {
		ub.ActiveBoard = position
	} else {
		ub.ActiveBoard = -1
	}
	ub.updateGameState()
	if ub.CurrentTurn == X {
		ub.CurrentTurn = O
	} else {
		ub.CurrentTurn = X
	}
	return nil
}

func (ub *UltimateBoard) updateGameState() {
	metaBoard := [9]CellState{}
	for i, board := range ub.Boards {
		switch board.State {
		case XWins:
			metaBoard[i] = X
		case OWins:
			metaBoard[i] = O
		default:
			metaBoard[i] = Empty
		}
	}
	for i := 0; i < 3; i++ {
		if metaBoard[i*3] != Empty &&
			metaBoard[i*3] == metaBoard[i*3+1] &&
			metaBoard[i*3+1] == metaBoard[i*3+2] {
			if metaBoard[i*3] == X {
				ub.State = XWins
			} else {
				ub.State = OWins
			}
			return
		}
	}
	for i := 0; i < 3; i++ {
		if metaBoard[i] != Empty &&
			metaBoard[i] == metaBoard[i+3] &&
			metaBoard[i+3] == metaBoard[i+6] {
			if metaBoard[i] == X {
				ub.State = XWins
			} else {
				ub.State = OWins
			}
			return
		}
	}
	if metaBoard[0] != Empty && metaBoard[0] == metaBoard[4] && metaBoard[4] == metaBoard[8] {
		if metaBoard[0] == X {
			ub.State = XWins
		} else {
			ub.State = OWins
		}
		return
	}
	if metaBoard[2] != Empty && metaBoard[2] == metaBoard[4] && metaBoard[4] == metaBoard[6] {
		if metaBoard[2] == X {
			ub.State = XWins
		} else {
			ub.State = OWins
		}
		return
	}
	allDecided := true
	for _, board := range ub.Boards {
		if board.State == Undecided {
			allDecided = false
			break
		}
	}
	if allDecided && ub.State == Undecided {
		ub.State = Draw
	}
}

func (ub *UltimateBoard) GetAvailableBoards() []int {
	if ub.State != Undecided {
		return []int{}
	}
	if ub.ActiveBoard != -1 {
		if ub.Boards[ub.ActiveBoard].State == Undecided {
			return []int{ub.ActiveBoard}
		}
		ub.ActiveBoard = -1
	}
	available := []int{}
	for i, board := range ub.Boards {
		if board.State == Undecided {
			available = append(available, i)
		}
	}
	return available
}

func (ub *UltimateBoard) GetBoardDisplay() string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Current turn: %s\n", ub.CurrentTurn))
	if ub.ActiveBoard != -1 {
		result.WriteString(fmt.Sprintf("Must play on board: %c\n", 'A'+ub.ActiveBoard))
	} else {
		result.WriteString("Can play on any available board\n")
	}
	result.WriteString(fmt.Sprintf("Game state: %s\n\n", ub.getGameStateString()))
	result.WriteString("      | A |         | B |         | C | \n")
	result.WriteString("   -----------   -----------   -----------\n")
	for row := 0; row < 3; row++ {
		result.WriteString(fmt.Sprintf("    %s | %s | %s     %s | %s | %s     %s | %s | %s\n",
			ub.Boards[0].Cells[row*3].String(),
			ub.Boards[0].Cells[row*3+1].String(),
			ub.Boards[0].Cells[row*3+2].String(),
			ub.Boards[1].Cells[row*3].String(),
			ub.Boards[1].Cells[row*3+1].String(),
			ub.Boards[1].Cells[row*3+2].String(),
			ub.Boards[2].Cells[row*3].String(),
			ub.Boards[2].Cells[row*3+1].String(),
			ub.Boards[2].Cells[row*3+2].String()))
		if row < 2 {
			result.WriteString("   -----------   -----------   -----------\n")
		}
	}
	result.WriteString("\n")
	result.WriteString("   -----------   -----------   -----------\n")
	result.WriteString("\n")
	result.WriteString("      | D |         | E |         | F | \n")
	result.WriteString("   -----------   -----------   -----------\n")
	for row := 0; row < 3; row++ {
		result.WriteString(fmt.Sprintf("    %s | %s | %s     %s | %s | %s     %s | %s | %s\n",
			ub.Boards[3].Cells[row*3].String(),
			ub.Boards[3].Cells[row*3+1].String(),
			ub.Boards[3].Cells[row*3+2].String(),
			ub.Boards[4].Cells[row*3].String(),
			ub.Boards[4].Cells[row*3+1].String(),
			ub.Boards[4].Cells[row*3+2].String(),
			ub.Boards[5].Cells[row*3].String(),
			ub.Boards[5].Cells[row*3+1].String(),
			ub.Boards[5].Cells[row*3+2].String()))
		if row < 2 {
			result.WriteString("   -----------   -----------   -----------\n")
		}
	}
	result.WriteString("\n")
	result.WriteString("   -----------   -----------   -----------\n")
	result.WriteString("\n")
	result.WriteString("      | G |         | H |         | I | \n")
	result.WriteString("   -----------   -----------   -----------\n")
	for row := 0; row < 3; row++ {
		result.WriteString(fmt.Sprintf("    %s | %s | %s     %s | %s | %s     %s | %s | %s\n",
			ub.Boards[6].Cells[row*3].String(),
			ub.Boards[6].Cells[row*3+1].String(),
			ub.Boards[6].Cells[row*3+2].String(),
			ub.Boards[7].Cells[row*3].String(),
			ub.Boards[7].Cells[row*3+1].String(),
			ub.Boards[7].Cells[row*3+2].String(),
			ub.Boards[8].Cells[row*3].String(),
			ub.Boards[8].Cells[row*3+1].String(),
			ub.Boards[8].Cells[row*3+2].String()))
		if row < 2 {
			result.WriteString("   -----------   -----------   -----------\n")
		}
	}
	result.WriteString("\n")
	result.WriteString("Board States: ")
	result.WriteString(fmt.Sprintf("A:%s B:%s C:%s  ",
		ub.Boards[0].State.String(),
		ub.Boards[1].State.String(),
		ub.Boards[2].State.String()))
	result.WriteString(fmt.Sprintf("D:%s E:%s F:%s  ",
		ub.Boards[3].State.String(),
		ub.Boards[4].State.String(),
		ub.Boards[5].State.String()))
	result.WriteString(fmt.Sprintf("G:%s H:%s I:%s\n",
		ub.Boards[6].State.String(),
		ub.Boards[7].State.String(),
		ub.Boards[8].State.String()))
	result.WriteString("\n")
	result.WriteString("MOVE LEGEND - Each board has cells numbered 1-9:\n")
	result.WriteString("   1 | 2 | 3       Examples:\n")
	result.WriteString("   ---------       A1 = Board A, cell 1 (top-left)\n")
	result.WriteString("   4 | 5 | 6       E5 = Board E, cell 5 (center)\n")
	result.WriteString("   ---------       I9 = Board I, cell 9 (bottom-right)\n")
	result.WriteString("   7 | 8 | 9\n")
	result.WriteString("\nEnter your move (e.g., A5, D2, I9): ")
	return result.String()
}

func (ub *UltimateBoard) getGameStateString() string {
	switch ub.State {
	case XWins:
		return "X Wins!"
	case OWins:
		return "O Wins!"
	case Draw:
		return "Draw!"
	default:
		return "In Progress"
	}
}
