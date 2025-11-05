package ugn

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/eshahhh/ultimatetictactoe/internal/game"
)

type GameLogger struct {
	ugnGame     *UGNGame
	gamesDir    string
	gameStarted bool
}

func NewGameLogger(gamesDir string) *GameLogger {
	return &GameLogger{
		gamesDir: gamesDir,
	}
}

func (gl *GameLogger) StartGame(gameID, playerX, playerO string) error {
	if err := os.MkdirAll(gl.gamesDir, 0755); err != nil {
		return fmt.Errorf("failed to create games directory: %v", err)
	}
	gl.ugnGame = NewUGNGame(gameID, playerX, playerO)
	gl.gameStarted = true
	return nil
}

func (gl *GameLogger) LogMove(move *game.Move, board *game.UltimateBoard, beforeGameState game.BoardState, beforeSmallState game.BoardState) error {
	if !gl.gameStarted {
		return fmt.Errorf("game logging not started")
	}
	ugnMove := GenerateUGNMove(move, board, beforeGameState, beforeSmallState)
	gl.ugnGame.AddMove(*ugnMove)
	return nil
}

func (gl *GameLogger) EndGame(result string) error {
	return gl.EndGameWithComment(result, "")
}

func (gl *GameLogger) EndGameWithComment(result, comment string) error {
	if !gl.gameStarted {
		return fmt.Errorf("game logging not started")
	}
	gl.ugnGame.SetResult(result)
	if comment != "" {
		gl.ugnGame.SetComment(comment)
	}
	filename := gl.ugnGame.GenerateFilename()
	filepath := filepath.Join(gl.gamesDir, filename)
	err := gl.ugnGame.WriteUGNFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to save UGN file: %v", err)
	}
	gl.gameStarted = false
	return nil
}

func (gl *GameLogger) GetCurrentGame() *UGNGame {
	return gl.ugnGame
}

func (gl *GameLogger) IsGameStarted() bool {
	return gl.gameStarted
}
