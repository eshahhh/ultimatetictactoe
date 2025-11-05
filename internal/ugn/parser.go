package ugn

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/eshahhh/ultimatetictactoe/internal/game"
)

type UGNMove struct {
	BoardIndex int
	Position   int
	SmallWin   bool
	SmallDraw  bool
	GameDraw   bool
	GameWin    bool
}

type GameMetadata struct {
	GameID  string
	Date    string
	Time    string
	PlayerX string
	PlayerO string
	Result  string
	Comment string // comment for the game result (e.g., "X wins by resignation")
}

type UGNGame struct {
	Metadata GameMetadata
	Moves    []UGNMove
}

func ParseMove(moveStr string) (*UGNMove, error) {
	moveStr = strings.TrimSpace(strings.ToUpper(moveStr))
	re := regexp.MustCompile(`^([A-I])([1-9])([!/\%#]*)$`)
	matches := re.FindStringSubmatch(moveStr)
	if len(matches) < 3 {
		return nil, fmt.Errorf("invalid UGN move format: %s", moveStr)
	}
	boardLetter := matches[1]
	boardIndex := int(boardLetter[0] - 'A')
	positionStr := matches[2]
	position, err := strconv.Atoi(positionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid position in UGN move: %s", positionStr)
	}
	position--
	symbols := ""
	if len(matches) > 3 {
		symbols = matches[3]
	}
	move := &UGNMove{
		BoardIndex: boardIndex,
		Position:   position,
		SmallWin:   strings.Contains(symbols, "!"),
		SmallDraw:  strings.Contains(symbols, "/"),
		GameDraw:   strings.Contains(symbols, "%"),
		GameWin:    strings.Contains(symbols, "#"),
	}
	return move, nil
}

func (m *UGNMove) ToString() string {
	boardLetter := string(rune('A' + m.BoardIndex))
	position := m.Position + 1
	symbols := ""
	if m.SmallWin {
		symbols += "!"
	}
	if m.SmallDraw {
		symbols += "/"
	}
	if m.GameDraw {
		symbols += "%"
	}
	if m.GameWin {
		symbols += "#"
	}
	return fmt.Sprintf("%s%d%s", boardLetter, position, symbols)
}

func GenerateUGNMove(move *game.Move, board *game.UltimateBoard, beforeState game.BoardState, beforeSmallState game.BoardState) *UGNMove {
	ugnMove := &UGNMove{
		BoardIndex: move.BoardIndex,
		Position:   move.Position,
	}
	currentSmallState := board.Boards[move.BoardIndex].State
	if beforeSmallState == game.Undecided && (currentSmallState == game.XWins || currentSmallState == game.OWins) {
		ugnMove.SmallWin = true
	}
	if beforeSmallState == game.Undecided && currentSmallState == game.Draw {
		ugnMove.SmallDraw = true
	}
	if beforeState == game.Undecided && (board.State == game.XWins || board.State == game.OWins) {
		ugnMove.GameWin = true
	}
	if beforeState == game.Undecided && board.State == game.Draw {
		ugnMove.GameDraw = true
	}
	return ugnMove
}

func ParseUGNFile(filename string) (*UGNGame, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open UGN file: %v", err)
	}
	defer file.Close()
	game := &UGNGame{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			break
		}
		re := regexp.MustCompile(`^\[(\w+)\s+"([^"]+)"\]$`)
		matches := re.FindStringSubmatch(line)
		if len(matches) == 3 {
			key := matches[1]
			value := matches[2]
			switch key {
			case "GameID":
				game.Metadata.GameID = value
			case "Date":
				game.Metadata.Date = value
			case "Time":
				game.Metadata.Time = value
			case "PlayerX":
				game.Metadata.PlayerX = value
			case "PlayerO":
				game.Metadata.PlayerO = value
			case "Result":
				game.Metadata.Result = value
			case "Comment":
				game.Metadata.Comment = value
			}
		}
	}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Check if this is a result line (1-0, 0-1, 1/2-1/2, or *)
		if line == "1-0" || line == "0-1" || line == "1/2-1/2" || line == "*" {
			// We can skip it as the result is already in metadata
			continue
		}

		moveStrings := strings.Fields(line)
		for _, moveStr := range moveStrings {
			move, err := ParseMove(moveStr)
			if err != nil {
				return nil, fmt.Errorf("failed to parse move '%s': %v", moveStr, err)
			}
			game.Moves = append(game.Moves, *move)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading UGN file: %v", err)
	}
	return game, nil
}

func (g *UGNGame) WriteUGNFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create UGN file: %v", err)
	}
	defer file.Close()
	fmt.Fprintf(file, "[GameID \"%s\"]\n", g.Metadata.GameID)
	fmt.Fprintf(file, "[Date \"%s\"]\n", g.Metadata.Date)
	fmt.Fprintf(file, "[Time \"%s\"]\n", g.Metadata.Time)
	fmt.Fprintf(file, "[PlayerX \"%s\"]\n", g.Metadata.PlayerX)
	fmt.Fprintf(file, "[PlayerO \"%s\"]\n", g.Metadata.PlayerO)
	fmt.Fprintf(file, "[Result \"%s\"]\n", g.Metadata.Result)
	if g.Metadata.Comment != "" {
		fmt.Fprintf(file, "[Comment \"%s\"]\n", g.Metadata.Comment)
	}
	fmt.Fprintf(file, "\n")
	for i, move := range g.Moves {
		if i > 0 && i%2 == 0 {
			fmt.Fprintf(file, "\n")
		} else if i > 0 {
			fmt.Fprintf(file, " ")
		}
		fmt.Fprintf(file, "%s", move.ToString())
	}
	fmt.Fprintf(file, "\n")

	var resultLine string
	switch g.Metadata.Result {
	case "X":
		resultLine = "1-0"
	case "O":
		resultLine = "0-1"
	case "Draw":
		resultLine = "1/2-1/2"
	default:
		resultLine = "*" // In progress or unknown result
	}
	fmt.Fprintf(file, "%s\n", resultLine)

	return nil
}

func (g *UGNGame) GenerateFilename() string {
	date := strings.ReplaceAll(g.Metadata.Date, "-", "")
	time := strings.ReplaceAll(g.Metadata.Time, ":", "")
	return fmt.Sprintf("%s_%s_%s.ugn", date, time, g.Metadata.GameID)
}

func NewUGNGame(gameID, playerX, playerO string) *UGNGame {
	now := time.Now()
	return &UGNGame{
		Metadata: GameMetadata{
			GameID:  gameID,
			Date:    now.Format("2006-01-02"),
			Time:    now.Format("15:04:05"),
			PlayerX: playerX,
			PlayerO: playerO,
			Result:  "In Progress",
		},
		Moves: make([]UGNMove, 0),
	}
}

func (g *UGNGame) AddMove(move UGNMove) {
	g.Moves = append(g.Moves, move)
}

func (g *UGNGame) SetResult(result string) {
	g.Metadata.Result = result
}

func (g *UGNGame) SetComment(comment string) {
	g.Metadata.Comment = comment
}

func (g *UGNGame) GetMovesString() string {
	var moves []string
	for _, move := range g.Moves {
		moves = append(moves, move.ToString())
	}
	return strings.Join(moves, " ")
}
