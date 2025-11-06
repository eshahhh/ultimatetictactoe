package matchmaking

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

// First-come, first-served matchmaking
type SimpleMatchmaker struct {
	queue   []*PlayerRequest
	mutex   sync.RWMutex
	maxSize int
}

func NewSimpleMatchmaker(maxPlayersPerGame int) *SimpleMatchmaker {
	return &SimpleMatchmaker{
		queue:   make([]*PlayerRequest, 0),
		maxSize: maxPlayersPerGame,
	}
}

func (sm *SimpleMatchmaker) AddPlayer(player *PlayerRequest) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	for _, p := range sm.queue {
		if p.ID == player.ID {
			return fmt.Errorf("player %s is already in queue", player.ID)
		}
	}
	player.JoinedAt = time.Now()
	player.Mode = SimpleMode
	sm.queue = append(sm.queue, player)
	return nil
}

func (sm *SimpleMatchmaker) RemovePlayer(playerID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	for i, player := range sm.queue {
		if player.ID == playerID {
			sm.queue = append(sm.queue[:i], sm.queue[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("player %s not found in queue", playerID)
}

func (sm *SimpleMatchmaker) FindMatch() []*GameMatch {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	var matches []*GameMatch
	for len(sm.queue) >= sm.maxSize {
		matchedPlayers := make([]*PlayerRequest, sm.maxSize)
		copy(matchedPlayers, sm.queue[:sm.maxSize])
		sm.queue = sm.queue[sm.maxSize:]
		gameID := generateGameID()
		match := &GameMatch{
			GameID:    gameID,
			Players:   matchedPlayers,
			CreatedAt: time.Now(),
			Mode:      SimpleMode,
		}
		matches = append(matches, match)
	}
	return matches
}

func (sm *SimpleMatchmaker) GetQueueSize() int {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return len(sm.queue)
}

func (sm *SimpleMatchmaker) GetMode() MatchmakingMode {
	return SimpleMode
}

func (sm *SimpleMatchmaker) GetQueuedPlayers() []*PlayerRequest {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	players := make([]*PlayerRequest, len(sm.queue))
	copy(players, sm.queue)
	return players
}

func generateGameID() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const idLength = 8
	result := make([]byte, idLength)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}
	return string(result)
}
