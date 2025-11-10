package main

import (
	"crypto/rand"
	"encoding/json"
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
		return true
	},
}

type GameServer struct {
	gameManager    *game.GameManager
	matchmaker     *matchmaking.MatchmakingManager
	gamesDir       string
	playerSessions map[string]*websocket.Conn
}

func NewGameServer() *GameServer {
	gs := &GameServer{
		gameManager:    game.NewGameManager(),
		gamesDir:       "games",
		playerSessions: make(map[string]*websocket.Conn),
	}

	gs.matchmaker = matchmaking.NewMatchmakingManager(gs.onMatchFound)
	gs.matchmaker.Start()

	return gs
}

func sendJSONMessage(conn *websocket.Conn, msgType game.MessageType, payload interface{}) error {
	msg := game.WebSocketMessage{
		Type:    msgType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

func sendGameStateToPlayer(session *game.GameSession, player *game.Player) error {
	gameState := session.GetGameStateForPlayer(player)
	log.Printf("Sending game state to %s. UGN moves count: %d", player.Name, len(gameState.UGNMoves))
	if len(gameState.UGNMoves) > 0 {
		log.Printf("UGN moves: %v", gameState.UGNMoves)
	}
	return sendJSONMessage(player.Conn, game.MessageTypeGameState, gameState)
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

	sessionLogger := ugn.NewGameLogger(gs.gamesDir)
	session.SetLogger(sessionLogger)

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
		} else {
			log.Printf("Game logging started for game %s", match.GameID)
		}
	}

	for _, player := range session.Players {
		err := sendGameStateToPlayer(session, player)
		if err != nil {
			log.Printf("Failed to send game state to player %s: %v", player.Name, err)
		}

		welcomeMsg := fmt.Sprintf("Match found! Game ID: %s. You are player %s", match.GameID, player.Symbol)
		sendJSONMessage(player.Conn, game.MessageTypeInfo, game.InfoPayload{Message: welcomeMsg})
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

	sendJSONMessage(conn, game.MessageTypeWelcome, game.WelcomePayload{
		PlayerID:   playerID,
		PlayerName: playerName,
		Message:    fmt.Sprintf("Welcome %s! Finding you a match...", playerName),
	})

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
		sendJSONMessage(conn, game.MessageTypeError, game.ErrorPayload{Message: err.Error()})
		return
	}

	queueMsg := fmt.Sprintf("You're in the matchmaking queue. Players waiting: %d", gs.matchmaker.GetTotalQueueSize())
	sendJSONMessage(conn, game.MessageTypeInfo, game.InfoPayload{Message: queueMsg})

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
				sendJSONMessage(conn, game.MessageTypeInfo, game.InfoPayload{Message: "Goodbye!"})
				break
			}

			if moveStr == "status" {
				if currentSession != nil {
					sendGameStateToPlayer(currentSession, currentPlayer)
				} else {
					statusMsg := fmt.Sprintf("Waiting for match... Players in queue: %d", gs.matchmaker.GetTotalQueueSize())
					sendJSONMessage(conn, game.MessageTypeInfo, game.InfoPayload{Message: statusMsg})
				}
				continue
			}

			if moveStr == "help" || moveStr == "?" {
				helpMsg := "Commands:\n" +
					"  A1-I9: Make a move (e.g., A1, B5, I9)\n" +
					"  R or resign: Resign from the game\n" +
					"  board/show: Request board update\n" +
					"  status: Show game status\n" +
					"  quit/exit: Leave the game"
				sendJSONMessage(conn, game.MessageTypeInfo, game.InfoPayload{Message: helpMsg})
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
					sendJSONMessage(conn, game.MessageTypeInfo, game.InfoPayload{
						Message: "Still waiting for a match... Type 'status' for queue info",
					})
					continue
				}
			}

			if moveStr == "board" || moveStr == "show" {
				sendGameStateToPlayer(currentSession, currentPlayer)
				continue
			}

			if game.IsResignation(moveStr) {
				err := currentSession.ResignGame(currentPlayer)
				if err != nil {
					sendJSONMessage(conn, game.MessageTypeError, game.ErrorPayload{Message: "Cannot resign: " + err.Error()})
					continue
				}

				resignMsg := fmt.Sprintf("Player %s (%s) has resigned!", playerName, currentPlayer.Symbol)
				for _, player := range currentSession.Players {
					if player != nil {
						sendJSONMessage(player.Conn, game.MessageTypeInfo, game.InfoPayload{Message: resignMsg})
						sendGameStateToPlayer(currentSession, player)
					}
				}

				var winnerName string
				if currentSession.Winner == game.X {
					winnerName = currentSession.Players[0].Name
					if currentSession.Players[0].Symbol == game.O {
						winnerName = currentSession.Players[1].Name
					}
				} else {
					winnerName = currentSession.Players[1].Name
					if currentSession.Players[1].Symbol == game.X {
						winnerName = currentSession.Players[0].Name
					}
				}

				gameOverPayload := game.GameOverPayload{
					Winner:     currentSession.Winner.String(),
					WinnerName: winnerName,
					Message:    fmt.Sprintf("%s wins by resignation!", winnerName),
					Comment:    "resignation",
				}

				for _, player := range currentSession.Players {
					if player != nil {
						sendJSONMessage(player.Conn, game.MessageTypeGameOver, gameOverPayload)
					}
				}
				continue
			}

			if moveStr == "DRAW" {
				err := currentSession.OfferDraw(currentPlayer)
				if err != nil {
					sendJSONMessage(conn, game.MessageTypeError, game.ErrorPayload{Message: "Cannot offer draw: " + err.Error()})
					continue
				}

				opponent := currentSession.GetOpponent(currentPlayer)
				if opponent != nil {
					drawOfferMsg := fmt.Sprintf("Player %s has offered a draw. Type ACCEPT_DRAW or DECLINE_DRAW", playerName)
					sendJSONMessage(opponent.Conn, game.MessageTypeDrawOffer, game.DrawOfferPayload{
						OfferedBy: playerName,
						Message:   drawOfferMsg,
					})
				}
				sendJSONMessage(conn, game.MessageTypeInfo, game.InfoPayload{Message: "Draw offer sent"})
				continue
			}

			if moveStr == "ACCEPT_DRAW" {
				if !currentSession.DrawOfferPending {
					sendJSONMessage(conn, game.MessageTypeError, game.ErrorPayload{Message: "No draw offer pending"})
					continue
				}

				if currentSession.DrawOfferedBy == currentPlayer {
					sendJSONMessage(conn, game.MessageTypeError, game.ErrorPayload{Message: "Cannot accept your own draw offer"})
					continue
				}

				err := currentSession.AcceptDraw()
				if err != nil {
					sendJSONMessage(conn, game.MessageTypeError, game.ErrorPayload{Message: "Error accepting draw: " + err.Error()})
					continue
				}

				for _, player := range currentSession.Players {
					if player != nil {
						sendJSONMessage(player.Conn, game.MessageTypeInfo, game.InfoPayload{Message: "Draw offer accepted! Game ended in a draw."})
						sendGameStateToPlayer(currentSession, player)
					}
				}

				gameOverPayload := game.GameOverPayload{
					Winner:     "Draw",
					WinnerName: "Draw",
					Message:    "Game ended in a draw by agreement",
					Comment:    "agreement",
				}

				for _, player := range currentSession.Players {
					if player != nil {
						sendJSONMessage(player.Conn, game.MessageTypeGameOver, gameOverPayload)
					}
				}
				continue
			}

			if moveStr == "DECLINE_DRAW" {
				if !currentSession.DrawOfferPending {
					sendJSONMessage(conn, game.MessageTypeError, game.ErrorPayload{Message: "No draw offer pending"})
					continue
				}

				if currentSession.DrawOfferedBy == currentPlayer {
					sendJSONMessage(conn, game.MessageTypeError, game.ErrorPayload{Message: "Cannot decline your own draw offer"})
					continue
				}

				err := currentSession.DeclineDraw()
				if err != nil {
					sendJSONMessage(conn, game.MessageTypeError, game.ErrorPayload{Message: "Error declining draw: " + err.Error()})
					continue
				}

				opponent := currentSession.GetOpponent(currentPlayer)
				if opponent != nil {
					sendJSONMessage(opponent.Conn, game.MessageTypeInfo, game.InfoPayload{Message: "Draw offer declined"})
				}
				sendJSONMessage(conn, game.MessageTypeInfo, game.InfoPayload{Message: "Draw offer declined"})
				continue
			}

			move, err := game.ParseMove(moveStr)
			if err != nil {
				sendJSONMessage(conn, game.MessageTypeError, game.ErrorPayload{Message: "Invalid move format: " + err.Error()})
				continue
			}

			err = currentSession.MakeMove(currentPlayer, move)
			if err != nil {
				sendJSONMessage(conn, game.MessageTypeError, game.ErrorPayload{Message: "Invalid move: " + err.Error()})
				continue
			}

			log.Printf("Move made successfully. Current UGN moves: %v", currentSession.GetUGNMoves())

			movePayload := game.MovePayload{
				PlayerName:   playerName,
				PlayerSymbol: currentPlayer.Symbol.String(),
				Move:         moveStr,
				BoardIndex:   move.BoardIndex,
				Position:     move.Position,
			}

			for _, player := range currentSession.Players {
				if player != nil {
					sendJSONMessage(player.Conn, game.MessageTypeMove, movePayload)
					sendGameStateToPlayer(currentSession, player)
				}
			}

			if currentSession.Finished {
				var winnerName string
				var winner string

				switch currentSession.Winner {
				case game.X:
					winner = "X"
					for _, p := range currentSession.Players {
						if p != nil && p.Symbol == game.X {
							winnerName = p.Name
							break
						}
					}
				case game.O:
					winner = "O"
					for _, p := range currentSession.Players {
						if p != nil && p.Symbol == game.O {
							winnerName = p.Name
							break
						}
					}
				default:
					winner = "Draw"
					winnerName = "Draw"
				}

				gameOverPayload := game.GameOverPayload{
					Winner:     winner,
					WinnerName: winnerName,
					Message:    fmt.Sprintf("Game Over - %s!", currentSession.GetGameStatus()),
					Comment:    "",
				}

				for _, player := range currentSession.Players {
					if player != nil {
						sendJSONMessage(player.Conn, game.MessageTypeGameOver, gameOverPayload)
					}
				}
			}
		}
	}

	gs.matchmaker.RemovePlayer(playerID)

	if currentSession != nil && currentPlayer != nil {
		currentSession.RemovePlayer(conn)
		log.Printf("Player %s (%s) disconnected from game %s", playerName, playerID, currentGameID)

		if opponent := currentSession.GetOpponent(currentPlayer); opponent != nil {
			sendJSONMessage(opponent.Conn, game.MessageTypeInfo, game.InfoPayload{
				Message: fmt.Sprintf("Player %s has disconnected", playerName),
			})
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

Connect to: ws://localhost:39171/ws
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

	log.Println("Ultimate Tic-Tac-Toe Server with Matchmaking starting on :39171")
	log.Println("WebSocket endpoint: ws://localhost:39171/ws")
	log.Println("Optional query parameter: ?name=YourName")
	log.Println("Matchmaking system: ACTIVE")

	log.Fatal(http.ListenAndServe(":39171", nil))
}
