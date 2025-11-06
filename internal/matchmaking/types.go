package matchmaking

import (
	"time"

	"github.com/gorilla/websocket"
)

type MatchmakingMode int

const (
	SimpleMode MatchmakingMode = iota // First-come, first-served
	EloMode                           // ELO-based matchmaking
	RankedMode                        // Ranked matchmaking
	CustomMode                        // Custom game modes
)

// Player looking for a match
type PlayerRequest struct {
	ID         string          // Unique player ID
	Name       string          // Player display name
	Connection *websocket.Conn // WebSocket connection
	JoinedAt   time.Time       // When player joined the queue
	Mode       MatchmakingMode // Matchmaking mode preference
	// EloRating  int             // Player's ELO rating
	// Preferences map[string]interface{} // Preferences (e.g. X or O)
}

// Match between players
type GameMatch struct {
	GameID    string           // Unique game identifier
	Players   []*PlayerRequest // Matched players
	CreatedAt time.Time        // When the match was created
	Mode      MatchmakingMode  // Matchmaking mode used
}

type Matchmaker interface {
	// Add a player to the matchmaking queue
	AddPlayer(player *PlayerRequest) error

	// Remove a player from the matchmaking queue
	RemovePlayer(playerID string) error

	// Find match for waiting players
	FindMatch() []*GameMatch

	// Current people searching
	GetQueueSize() int

	// Matchmaking mode
	GetMode() MatchmakingMode
}

// Matchmaking attempt result
type MatchmakingResult struct {
	Success bool       // Successful or not
	Match   *GameMatch // Match (if successful)
	Error   error      // Error (if not)
}

// Match is found
type MatchmakingCallback func(match *GameMatch) error
