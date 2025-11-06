package matchmaking

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type MatchmakingManager struct {
	matchmakers      map[MatchmakingMode]Matchmaker
	matchCallback    MatchmakingCallback
	tickerInterval   time.Duration
	stopChan         chan struct{}
	mutex            sync.RWMutex
	running          bool
}

func NewMatchmakingManager(callback MatchmakingCallback) *MatchmakingManager {
	mm := &MatchmakingManager{
		matchmakers:    make(map[MatchmakingMode]Matchmaker),
		matchCallback:  callback,
		tickerInterval: 1 * time.Second,
		stopChan:       make(chan struct{}),
	}
	mm.RegisterMatchmaker(NewSimpleMatchmaker(2))
	return mm
}

func (mm *MatchmakingManager) RegisterMatchmaker(matchmaker Matchmaker) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	mm.matchmakers[matchmaker.GetMode()] = matchmaker
	log.Printf("Registered matchmaker for mode: %v", matchmaker.GetMode())
}

func (mm *MatchmakingManager) AddPlayer(player *PlayerRequest) error {
	mm.mutex.RLock()
	matchmaker, exists := mm.matchmakers[player.Mode]
	mm.mutex.RUnlock()
	if !exists {
		mm.mutex.RLock()
		matchmaker = mm.matchmakers[SimpleMode]
		mm.mutex.RUnlock()
		if matchmaker == nil {
			return fmt.Errorf("no matchmaker available for mode %v", player.Mode)
		}
		player.Mode = SimpleMode
	}
	err := matchmaker.AddPlayer(player)
	if err != nil {
		return fmt.Errorf("failed to add player to matchmaking: %w", err)
	}
	log.Printf("Player %s (%s) added to %v matchmaking queue", player.ID, player.Name, player.Mode)
	return nil
}

func (mm *MatchmakingManager) RemovePlayer(playerID string) error {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	var lastError error
	removed := false
	for mode, matchmaker := range mm.matchmakers {
		err := matchmaker.RemovePlayer(playerID)
		if err == nil {
			removed = true
			log.Printf("Player %s removed from %v matchmaking queue", playerID, mode)
		} else {
			lastError = err
		}
	}
	if !removed {
		return fmt.Errorf("player %s not found in any matchmaking queue: %w", playerID, lastError)
	}
	return nil
}

func (mm *MatchmakingManager) Start() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	if mm.running {
		log.Println("Matchmaking manager is already running")
		return
	}
	mm.running = true
	go mm.matchmakingLoop()
	log.Println("Matchmaking manager started")
}

func (mm *MatchmakingManager) Stop() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	if !mm.running {
		log.Println("Matchmaking manager is not running")
		return
	}
	mm.running = false
	close(mm.stopChan)
	log.Println("Matchmaking manager stopped")
}

func (mm *MatchmakingManager) matchmakingLoop() {
	ticker := time.NewTicker(mm.tickerInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			mm.processPendingMatches()
		case <-mm.stopChan:
			log.Println("Matchmaking loop stopped")
			return
		}
	}
}

func (mm *MatchmakingManager) processPendingMatches() {
	mm.mutex.RLock()
	matchmakers := make([]Matchmaker, 0, len(mm.matchmakers))
	for _, matchmaker := range mm.matchmakers {
		matchmakers = append(matchmakers, matchmaker)
	}
	mm.mutex.RUnlock()
	for _, matchmaker := range matchmakers {
		matches := matchmaker.FindMatch()
		for _, match := range matches {
			if mm.matchCallback != nil {
				err := mm.matchCallback(match)
				if err != nil {
					log.Printf("Error processing match %s: %v", match.GameID, err)
				} else {
					log.Printf("Match created: %s with %d players", match.GameID, len(match.Players))
				}
			}
		}
	}
}

func (mm *MatchmakingManager) GetQueueStatus() map[MatchmakingMode]int {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()
	status := make(map[MatchmakingMode]int)
	for mode, matchmaker := range mm.matchmakers {
		status[mode] = matchmaker.GetQueueSize()
	}
	return status
}

func (mm *MatchmakingManager) GetTotalQueueSize() int {
	total := 0
	for _, size := range mm.GetQueueStatus() {
		total += size
	}
	return total
}

func (mm *MatchmakingManager) SetTickerInterval(interval time.Duration) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	mm.tickerInterval = interval
	log.Printf("Matchmaking ticker interval set to %v", interval)
}
