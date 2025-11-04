package game

import (
	"testing"
)

func TestParseMove(t *testing.T) {
	tests := []struct {
		input       string
		expectError bool
		boardIndex  int
		position    int
	}{
		{"A1", false, 0, 0},
		{"A9", false, 0, 8},
		{"I9", false, 8, 8},
		{"E5", false, 4, 4},
		{"a1", false, 0, 0},   // lowercase
		{" B2 ", false, 1, 1}, // whitespace
		{"Z1", true, 0, 0},    // Invalid board
		{"A0", true, 0, 0},    // Invalid position
		{"AA", true, 0, 0},    // Invalid format
		{"123", true, 0, 0},   // Invalid format
		{"", true, 0, 0},      // Empty string
	}

	for _, test := range tests {
		move, err := ParseMove(test.input)

		if test.expectError {
			if err == nil {
				t.Errorf("Expected error for input '%s', but got none", test.input)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input '%s': %v", test.input, err)
			} else {
				if move.BoardIndex != test.boardIndex {
					t.Errorf("For input '%s', expected board index %d, got %d",
						test.input, test.boardIndex, move.BoardIndex)
				}
				if move.Position != test.position {
					t.Errorf("For input '%s', expected position %d, got %d",
						test.input, test.position, move.Position)
				}
			}
		}
	}
}

func TestSmallBoardWin(t *testing.T) {
	board := NewSmallBoard()

	board.MakeMove(0, X)
	board.MakeMove(1, X)
	board.MakeMove(2, X)

	if board.State != XWins {
		t.Errorf("Expected X to win, but board state is %v", board.State)
	}
}

func TestSmallBoardDraw(t *testing.T) {
	board := NewSmallBoard()

	// Create a draw situation:
	// X O X
	// O O X
	// O X O
	moves := []struct {
		pos    int
		player CellState
	}{
		{0, X}, {1, O}, {2, X},
		{3, O}, {4, O}, {5, X},
		{6, O}, {7, X}, {8, O},
	}

	for _, move := range moves {
		board.MakeMove(move.pos, move.player)
	}

	if board.State != Draw {
		t.Errorf("Expected draw, but board state is %v", board.State)
	}
}
func TestUltimateBoardActiveBoard(t *testing.T) {
	board := NewUltimateBoard()

	err := board.MakeMove(0, 4)
	if err != nil {
		t.Errorf("First move failed: %v", err)
	}

	if board.ActiveBoard != 4 {
		t.Errorf("Expected active board to be 4, got %d", board.ActiveBoard)
	}

	err = board.MakeMove(0, 0)
	if err == nil {
		t.Errorf("Should not be able to play on board 0 when active board is 4")
	}

	err = board.MakeMove(4, 0)
	if err != nil {
		t.Errorf("Move on active board failed: %v", err)
	}
}

func TestUltimateBoardWin(t *testing.T) {
	board := NewUltimateBoard()

	board.Boards[0].MakeMove(0, X)
	board.Boards[0].MakeMove(1, X)
	board.Boards[0].MakeMove(2, X)

	board.Boards[1].MakeMove(0, X)
	board.Boards[1].MakeMove(1, X)
	board.Boards[1].MakeMove(2, X)

	board.Boards[2].MakeMove(0, X)
	board.Boards[2].MakeMove(1, X)
	board.Boards[2].MakeMove(2, X)

	board.updateGameState()

	if board.State != XWins {
		t.Errorf("Expected X to win overall game, but state is %v", board.State)
	}
}

func TestValidateMoveFormat(t *testing.T) {
	validMoves := []string{"A1", "I9", "E5", "a1", " B2 "}
	invalidMoves := []string{"Z1", "A0", "AA", "123", "", "A10"}

	for _, move := range validMoves {
		if !ValidateMoveFormat(move) {
			t.Errorf("Move '%s' should be valid", move)
		}
	}

	for _, move := range invalidMoves {
		if ValidateMoveFormat(move) {
			t.Errorf("Move '%s' should be invalid", move)
		}
	}
}
