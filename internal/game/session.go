package game

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Player struct {
	Conn     *websocket.Conn
	Symbol   CellState
	Name     string
	LastSeen time.Time
}

type GameSession struct {
	ID       string
	Board    *UltimateBoard
	Players  [2]*Player
	Started  bool
	Finished bool
	Winner   CellState
	Logger   GameLogger
	mutex    sync.RWMutex
}

type GameLogger interface {
	StartGame(gameID, playerX, playerO string) error
	LogMove(move *Move, board *UltimateBoard, beforeGameState BoardState, beforeSmallState BoardState) error
	EndGame(result string) error
	EndGameWithComment(result, comment string) error
	IsGameStarted() bool
}

func NewGameSession(id string) *GameSession {
	return &GameSession{
		ID:      id,
		Board:   NewUltimateBoard(),
		Started: false,
		Winner:  Empty,
		Logger:  nil,
	}
}

func NewGameSessionWithPlayers(id string, player1Conn *websocket.Conn, player1Name string, player2Conn *websocket.Conn, player2Name string) *GameSession {
	session := NewGameSession(id)

	// Randomly assign X and O
	player1Symbol, player2Symbol := randomPlayerAssignment()

	session.Players[0] = &Player{
		Conn:     player1Conn,
		Symbol:   player1Symbol,
		Name:     player1Name,
		LastSeen: time.Now(),
	}

	session.Players[1] = &Player{
		Conn:     player2Conn,
		Symbol:   player2Symbol,
		Name:     player2Name,
		LastSeen: time.Now(),
	}

	session.Started = true
	return session
}

func randomPlayerAssignment() (CellState, CellState) {
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(2))

	if randomNum.Int64() == 0 {
		return X, O // Player 1 gets X, Player 2 gets O
	}
	return O, X // Player 1 gets O, Player 2 gets X
}

func (gs *GameSession) AddPlayer(conn *websocket.Conn, name string) (*Player, error) {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	var symbol CellState
	var slot int = -1

	for i, player := range gs.Players {
		if player == nil {
			slot = i
			if i == 0 {
				symbol = X
			} else {
				symbol = O
			}
			break
		}
	}

	if slot == -1 {
		return nil, fmt.Errorf("game session is full")
	}

	player := &Player{
		Conn:     conn,
		Symbol:   symbol,
		Name:     name,
		LastSeen: time.Now(),
	}

	gs.Players[slot] = player

	if gs.Players[0] != nil && gs.Players[1] != nil && !gs.Started {
		gs.Started = true

		if gs.Logger != nil {
			playerXName := gs.Players[0].Name
			playerOName := gs.Players[1].Name
			err := gs.Logger.StartGame(gs.ID, playerXName, playerOName)
			if err != nil {
				fmt.Printf("Failed to log game start: %v\n", err)
			}
		}
	}

	return player, nil
}

func (gs *GameSession) RemovePlayer(conn *websocket.Conn) {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	for i, player := range gs.Players {
		if player != nil && player.Conn == conn {
			gs.Players[i] = nil
			break
		}
	}
}

func (gs *GameSession) GetCurrentPlayer() *Player {
	gs.mutex.RLock()
	defer gs.mutex.RUnlock()

	if !gs.Started {
		return nil
	}

	for _, player := range gs.Players {
		if player != nil && player.Symbol == gs.Board.CurrentTurn {
			return player
		}
	}

	return nil
}

func (gs *GameSession) GetOpponent(currentPlayer *Player) *Player {
	gs.mutex.RLock()
	defer gs.mutex.RUnlock()

	for _, player := range gs.Players {
		if player != nil && player != currentPlayer {
			return player
		}
	}

	return nil
}

func (gs *GameSession) MakeMove(player *Player, move *Move) error {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	if !gs.Started {
		return fmt.Errorf("game has not started yet")
	}

	if gs.Finished {
		return fmt.Errorf("game is already finished")
	}

	if gs.Board.CurrentTurn != player.Symbol {
		return fmt.Errorf("it's not your turn")
	}

	beforeGameState := gs.Board.State
	beforeSmallState := gs.Board.Boards[move.BoardIndex].State

	err := gs.Board.MakeMove(move.BoardIndex, move.Position)
	if err != nil {
		return err
	}

	if gs.Logger != nil && gs.Logger.IsGameStarted() {
		err := gs.Logger.LogMove(move, gs.Board, beforeGameState, beforeSmallState)
		if err != nil {
		}
	}

	if gs.Board.State != Undecided {
		gs.Finished = true
		switch gs.Board.State {
		case XWins:
			gs.Winner = X
		case OWins:
			gs.Winner = O
		case Draw:
			gs.Winner = Empty
		}

		if gs.Logger != nil && gs.Logger.IsGameStarted() {
			var result string
			switch gs.Winner {
			case X:
				result = "X"
			case O:
				result = "O"
			default:
				result = "Draw"
			}
			err := gs.Logger.EndGame(result)
			if err != nil {
			}
		}
	}

	return nil
}

func (gs *GameSession) ResignGame(player *Player) error {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()

	if !gs.Started {
		return fmt.Errorf("game has not started yet")
	}

	if gs.Finished {
		return fmt.Errorf("game is already finished")
	}

	gs.Finished = true
	if player.Symbol == X {
		gs.Winner = O
	} else {
		gs.Winner = X
	}

	if gs.Logger != nil && gs.Logger.IsGameStarted() {
		var result string
		var comment string
		if gs.Winner == X {
			result = "X"
			comment = "X wins by resignation"
		} else {
			result = "O"
			comment = "O wins by resignation"
		}
		err := gs.Logger.EndGameWithComment(result, comment)
		if err != nil {
		}
	}

	return nil
}

func (gs *GameSession) GetGameStatus() string {
	gs.mutex.RLock()
	defer gs.mutex.RUnlock()

	if !gs.Started {
		playerCount := 0
		for _, player := range gs.Players {
			if player != nil {
				playerCount++
			}
		}
		return fmt.Sprintf("Waiting for players (%d/2 connected)", playerCount)
	}

	if gs.Finished {
		switch gs.Winner {
		case X:
			return "Game Over - X Wins!"
		case O:
			return "Game Over - O Wins!"
		default:
			return "Game Over - Draw!"
		}
	}

	currentPlayer := gs.GetCurrentPlayer()
	if currentPlayer != nil {
		return fmt.Sprintf("Game in progress - %s's turn (%s)",
			currentPlayer.Name, currentPlayer.Symbol)
	}

	return "Game in progress"
}

func (gs *GameSession) BroadcastToAll(message string) {
	gs.mutex.RLock()
	defer gs.mutex.RUnlock()

	for _, player := range gs.Players {
		if player != nil {
			player.Conn.WriteMessage(websocket.TextMessage, []byte(message))
		}
	}
}

func (gs *GameSession) SendToPlayer(player *Player, message string) {
	if player != nil && player.Conn != nil {
		player.Conn.WriteMessage(websocket.TextMessage, []byte(message))
	}
}

func (gs *GameSession) SetLogger(logger GameLogger) {
	gs.mutex.Lock()
	defer gs.mutex.Unlock()
	gs.Logger = logger
}

type GameManager struct {
	sessions map[string]*GameSession
	mutex    sync.RWMutex
}

func NewGameManager() *GameManager {
	return &GameManager{
		sessions: make(map[string]*GameSession),
	}
}

func (gm *GameManager) CreateSession(id string) *GameSession {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	session := NewGameSession(id)
	gm.sessions[id] = session

	return session
}

func (gm *GameManager) GetSession(id string) *GameSession {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()

	return gm.sessions[id]
}

func (gm *GameManager) GetOrCreateSession(id string) *GameSession {
	session := gm.GetSession(id)
	if session == nil {
		session = gm.CreateSession(id)
	}
	return session
}

func (gm *GameManager) RemoveSession(id string) {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	delete(gm.sessions, id)
}

func (gm *GameManager) GetActiveSessions() []string {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()

	sessions := make([]string, 0, len(gm.sessions))
	for id := range gm.sessions {
		sessions = append(sessions, id)
	}

	return sessions
}

func (gm *GameManager) AddSession(session *GameSession) {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()
	gm.sessions[session.ID] = session
}
