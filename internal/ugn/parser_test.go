package ugn

import (
	"testing"

	"github.com/eshahhh/ultimatetictactoe/internal/game"
)

func TestParseMove(t *testing.T) {
	tests := []struct {
		input    string
		expected UGNMove
		hasError bool
	}{
		{
			input: "A5",
			expected: UGNMove{
				BoardIndex: 0,
				Position:   4,
				SmallWin:   false,
				SmallDraw:  false,
				GameDraw:   false,
				GameWin:    false,
			},
			hasError: false,
		},
		{
			input: "E9!",
			expected: UGNMove{
				BoardIndex: 4,
				Position:   8,
				SmallWin:   true,
				SmallDraw:  false,
				GameDraw:   false,
				GameWin:    false,
			},
			hasError: false,
		},
		{
			input: "C6/",
			expected: UGNMove{
				BoardIndex: 2,
				Position:   5,
				SmallWin:   false,
				SmallDraw:  true,
				GameDraw:   false,
				GameWin:    false,
			},
			hasError: false,
		},
		{
			input: "D2#",
			expected: UGNMove{
				BoardIndex: 3,
				Position:   1,
				SmallWin:   false,
				SmallDraw:  false,
				GameDraw:   false,
				GameWin:    true,
			},
			hasError: false,
		},
		{
			input: "I5/%",
			expected: UGNMove{
				BoardIndex: 8,
				Position:   4,
				SmallWin:   false,
				SmallDraw:  true,
				GameDraw:   true,
				GameWin:    false,
			},
			hasError: false,
		},
		{
			input:    "J5", // Invalid board
			expected: UGNMove{},
			hasError: true,
		},
		{
			input:    "A0", // Invalid position
			expected: UGNMove{},
			hasError: true,
		},
	}

	for _, test := range tests {
		result, err := ParseMove(test.input)

		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for input '%s', but got none", test.input)
			}
			continue
		}

		if err != nil {
			t.Errorf("Unexpected error for input '%s': %v", test.input, err)
			continue
		}

		if *result != test.expected {
			t.Errorf("For input '%s', expected %+v, got %+v", test.input, test.expected, *result)
		}
	}
}

func TestUGNMoveToString(t *testing.T) {
	tests := []struct {
		move     UGNMove
		expected string
	}{
		{
			move: UGNMove{
				BoardIndex: 0,
				Position:   4,
			},
			expected: "A5",
		},
		{
			move: UGNMove{
				BoardIndex: 4,
				Position:   8,
				SmallWin:   true,
			},
			expected: "E9!",
		},
		{
			move: UGNMove{
				BoardIndex: 2,
				Position:   5,
				SmallDraw:  true,
			},
			expected: "C6/",
		},
		{
			move: UGNMove{
				BoardIndex: 3,
				Position:   1,
				GameWin:    true,
			},
			expected: "D2#",
		},
		{
			move: UGNMove{
				BoardIndex: 8,
				Position:   4,
				SmallDraw:  true,
				GameDraw:   true,
			},
			expected: "I5/%",
		},
	}

	for _, test := range tests {
		result := test.move.ToString()
		if result != test.expected {
			t.Errorf("For move %+v, expected '%s', got '%s'", test.move, test.expected, result)
		}
	}
}

func TestGenerateUGNMove(t *testing.T) {
	board := game.NewUltimateBoard()
	move := &game.Move{BoardIndex: 0, Position: 4} // A5

	// Test normal move (no special outcomes)
	ugnMove := GenerateUGNMove(move, board, game.Undecided, game.Undecided)

	if ugnMove.BoardIndex != 0 {
		t.Errorf("Expected BoardIndex 0, got %d", ugnMove.BoardIndex)
	}

	if ugnMove.Position != 4 {
		t.Errorf("Expected Position 4, got %d", ugnMove.Position)
	}

	if ugnMove.SmallWin || ugnMove.SmallDraw || ugnMove.GameWin || ugnMove.GameDraw {
		t.Errorf("Expected no special outcomes for normal move, got %+v", ugnMove)
	}
}
