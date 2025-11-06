package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/eshahhh/ultimatetictactoe/internal/game"
	"github.com/eshahhh/ultimatetictactoe/internal/matchmaking"
	"github.com/eshahhh/ultimatetictactoe/internal/ugn"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all
	},
}

type GameServer struct {
	gameManager    *game.GameManager
	matchmaker     *matchmaking.MatchmakingManager
	ugnLogger      *ugn.GameLogger
	playerSessions map[string]*websocket.Conn // playerID -> connection
}

func NewGameServer() *GameServer {
	gs := &GameServer{
		gameManager:    game.NewGameManager(),
		ugnLogger:      ugn.NewGameLogger("games"),
		playerSessions: make(map[string]*websocket.Conn),
	}

	gs.matchmaker = matchmaking.NewMatchmakingManager(gs.onMatchFound)
	gs.matchmaker.Start()

	return gs
}

func (gs *GameServer) onMatchFound(match *matchmaking.GameMatch) error {
	if len(match.Players) != 2 {
		return fmt.Errorf("invalid match: expected 2 players, got %d", len(match.Players))
	}

	player1 := match.Players[0]
	player2 := match.Players[1]

	session := game.NewGameSessionWithPlayers(
		match.GameID,
		player1.Connection, player1.Name,
		player2.Connection, player2.Name,
	)

	session.SetLogger(gs.ugnLogger)

	gs.gameManager.AddSession(session)

	var playerXName, playerOName string
	for _, player := range session.Players {
		if player.Symbol == game.X {
			playerXName = player.Name
		} else {
			playerOName = player.Name
		}
	}

	if session.Logger != nil {
		err := session.Logger.StartGame(match.GameID, playerXName, playerOName)
		if err != nil {
			log.Printf("Failed to start game logging: %v", err)
		}
	}

	for _, player := range session.Players {
		welcomeMsg := fmt.Sprintf("Match found! Game ID: %s\nYou are player %s\n%s",
			match.GameID, player.Symbol, session.GetGameStatus())
		session.SendToPlayer(player, welcomeMsg)
		session.SendToPlayer(player, session.Board.GetBoardDisplay())
	}

	log.Printf("Game %s started with players %s (%s) and %s (%s)",
		match.GameID,
		session.Players[0].Name, session.Players[0].Symbol,
		session.Players[1].Name, session.Players[1].Symbol)

	return nil
}

func generatePlayerID() string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	const idLength = 12

	result := make([]byte, idLength)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}

	return string(result)
}

func (gs *GameServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	playerName := r.URL.Query().Get("name")
	if playerName == "" {
		playerName = r.RemoteAddr
	}

	playerID := generatePlayerID()

	gs.playerSessions[playerID] = conn
	defer delete(gs.playerSessions, playerID)

	welcomeMsg := fmt.Sprintf("Welcome %s! Finding you a match...\nPlayer ID: %s", playerName, playerID)
	conn.WriteMessage(websocket.TextMessage, []byte(welcomeMsg))

	log.Printf("Player %s (%s) connected from %s", playerName, playerID, r.RemoteAddr)

	playerRequest := &matchmaking.PlayerRequest{
		ID:         playerID,
		Name:       playerName,
		Connection: conn,
		Mode:       matchmaking.SimpleMode,
	}

	err = gs.matchmaker.AddPlayer(playerRequest)
	if err != nil {
		log.Printf("Error adding player to matchmaker: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		return
	}

	queueMsg := fmt.Sprintf("You're in the matchmaking queue. Players waiting: %d", gs.matchmaker.GetTotalQueueSize())
	conn.WriteMessage(websocket.TextMessage, []byte(queueMsg))

	var currentGameID string
	var currentSession *game.GameSession
	var currentPlayer *game.Player

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from %s: %v", playerName, err)
			break
		}

		if messageType == websocket.TextMessage {
			moveStr := strings.TrimSpace(string(message))
			log.Printf("Received message '%s' from %s (%s)", moveStr, playerName, playerID)

			if moveStr == "quit" || moveStr == "exit" {
				conn.WriteMessage(websocket.TextMessage, []byte("Goodbye!"))
				break
			}

			if moveStr == "status" {
				if currentSession != nil {
					conn.WriteMessage(websocket.TextMessage, []byte(currentSession.GetGameStatus()))
				} else {
					statusMsg := fmt.Sprintf("Waiting for match... Players in queue: %d", gs.matchmaker.GetTotalQueueSize())
					conn.WriteMessage(websocket.TextMessage, []byte(statusMsg))
				}
				continue
			}

			if moveStr == "help" || moveStr == "?" {
				helpMsg := "Commands:\n" +
					"  A1-I9: Make a move (e.g., A1, B5, I9)\n" +
					"  R or resign: Resign from the game\n" +
					"  board/show: Display the current board\n" +
					"  status: Show game status\n" +
					"  quit/exit: Leave the game"
				conn.WriteMessage(websocket.TextMessage, []byte(helpMsg))
				continue
			}

			if currentSession == nil {
				sessions := gs.gameManager.GetActiveSessions()
				for _, sessionID := range sessions {
					session := gs.gameManager.GetSession(sessionID)
					if session != nil {
						for _, player := range session.Players {
							if player != nil && player.Conn == conn {
								currentGameID = sessionID
								currentSession = session
								currentPlayer = player
								break
							}
						}
						if currentSession != nil {
							break
						}
					}
				}

				if currentSession == nil {
					conn.WriteMessage(websocket.TextMessage, []byte("Still waiting for a match... Type 'status' for queue info"))
					continue
				}
			}

			if moveStr == "board" || moveStr == "show" {
				conn.WriteMessage(websocket.TextMessage, []byte(currentSession.Board.GetBoardDisplay()))
				continue
			}

			if game.IsResignation(moveStr) {
				err := currentSession.ResignGame(currentPlayer)
				if err != nil {
					conn.WriteMessage(websocket.TextMessage, []byte("Cannot resign: "+err.Error()))
					continue
				}

				resignMsg := fmt.Sprintf("Player %s (%s) has resigned!", playerName, currentPlayer.Symbol)
				currentSession.BroadcastToAll(resignMsg)

				currentSession.BroadcastToAll(currentSession.GetGameStatus())

				var finalMsg string
				if currentSession.Winner == game.X {
					finalMsg = "Player X wins by resignation!"
				} else {
					finalMsg = "Player O wins by resignation!"
				}
				currentSession.BroadcastToAll(finalMsg)
				continue
			}

			move, err := game.ParseMove(moveStr)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte("Invalid move format: "+err.Error()))
				continue
			}

			err = currentSession.MakeMove(currentPlayer, move)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte("Invalid move: "+err.Error()))
				continue
			}

			moveMsg := fmt.Sprintf("Player %s (%s) played %s",
				playerName, currentPlayer.Symbol, moveStr)
			currentSession.BroadcastToAll(moveMsg)

			currentSession.BroadcastToAll(currentSession.Board.GetBoardDisplay())

			currentSession.BroadcastToAll(currentSession.GetGameStatus())

			if currentSession.Finished {
				var finalMsg string
				switch currentSession.Winner {
				case game.X:
					finalMsg = "Player X wins the Ultimate Tic-Tac-Toe!"
				case game.O:
					finalMsg = "Player O wins the Ultimate Tic-Tac-Toe!"
				default:
					finalMsg = "Game ended in a draw! Good game!"
				}
				currentSession.BroadcastToAll(finalMsg)
			}
		}
	}

	gs.matchmaker.RemovePlayer(playerID)

	if currentSession != nil && currentPlayer != nil {
		currentSession.RemovePlayer(conn)
		log.Printf("Player %s (%s) disconnected from game %s", playerName, playerID, currentGameID)

		if opponent := currentSession.GetOpponent(currentPlayer); opponent != nil {
			currentSession.SendToPlayer(opponent, fmt.Sprintf("Player %s has disconnected", playerName))
		}
	} else {
		log.Printf("Player %s (%s) disconnected while in matchmaking queue", playerName, playerID)
	}
}

func main() {
	gameServer := NewGameServer()

	defer gameServer.matchmaker.Stop()

	http.HandleFunc("/ws", gameServer.handleWebSocket)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `Ultimate Tic-Tac-Toe WebSocket Server with Matchmaking!

Connect to: ws://localhost:8080/ws
Optional query parameter: ?name=YourName

How it works:
1. Connect to the server
2. Wait for matchmaking to find you an opponent
3. Play Ultimate Tic-Tac-Toe with random X/O assignment
4. Games are automatically logged in UGN format

Game Commands:
- A1-I9: Make a move (e.g., A1, B5, I9)
- R or resign: Resign from the game  
- board/show: Display the current board
- status: Show game/queue status
- help: Show this help message
- quit/exit: Leave the game

Features:
- Automatic matchmaking
- Multiple simultaneous games
- Random player assignment (X/O)
- UGN game logging
- Resignation support`)
	})

	log.Println("Ultimate Tic-Tac-Toe Server with Matchmaking starting on :8080")
	log.Println("WebSocket endpoint: ws://localhost:8080/ws")
	log.Println("Optional query parameter: ?name=YourName")
	log.Println("Matchmaking system: ACTIVE")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
