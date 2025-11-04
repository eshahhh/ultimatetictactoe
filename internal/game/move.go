package game

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Move struct {
	BoardIndex int // 0-8 corresponding to A-I
	Position   int // 0-8 corresponding to positions 1-9
}

func ParseMove(moveStr string) (*Move, error) {
	moveStr = strings.TrimSpace(strings.ToUpper(moveStr))

	if moveStr == "R" || moveStr == "RESIGN" {
		return nil, fmt.Errorf("resignation")
	}

	re := regexp.MustCompile(`^([A-I])([1-9])$`)
	matches := re.FindStringSubmatch(moveStr)

	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid move format: %s (expected format: A1-I9)", moveStr)
	}

	boardLetter := matches[1]
	boardIndex := int(boardLetter[0] - 'A')

	positionStr := matches[2]
	position, err := strconv.Atoi(positionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid position: %s", positionStr)
	}
	position-- // 1-based to 0-based indexing

	return &Move{
		BoardIndex: boardIndex,
		Position:   position,
	}, nil
}

func (m *Move) ToString() string {
	boardLetter := string(rune('A' + m.BoardIndex))
	position := m.Position + 1
	return fmt.Sprintf("%s%d", boardLetter, position)
}

func ValidateMoveFormat(moveStr string) bool {
	moveStr = strings.TrimSpace(strings.ToUpper(moveStr))
	re := regexp.MustCompile(`^([A-I])([1-9])$`)
	return re.MatchString(moveStr)
}

func IsResignation(moveStr string) bool {
	moveStr = strings.TrimSpace(strings.ToUpper(moveStr))
	return moveStr == "R" || moveStr == "RESIGN"
}
