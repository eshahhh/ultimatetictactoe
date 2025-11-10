package game

type MessageType string

const (
	MessageTypeGameState MessageType = "game_state"
	MessageTypeMove      MessageType = "move"
	MessageTypeError     MessageType = "error"
	MessageTypeInfo      MessageType = "info"
	MessageTypeGameOver  MessageType = "game_over"
	MessageTypeDrawOffer MessageType = "draw_offer"
	MessageTypeWelcome   MessageType = "welcome"
)

type WebSocketMessage struct {
	Type    MessageType `json:"type"`
	Payload interface{} `json:"payload"`
}

type WelcomePayload struct {
	PlayerID   string `json:"player_id"`
	PlayerName string `json:"player_name"`
	Message    string `json:"message"`
}

type GameStatePayload struct {
	GameID      string         `json:"game_id"`
	Board       BoardStateData `json:"board"`
	CurrentTurn string         `json:"current_turn"` // "X" or "O"
	YourSymbol  string         `json:"your_symbol"`  // "X" or "O"
	ActiveBoard int            `json:"active_board"` // -1 for any, 0-8 for specific
	GameStatus  string         `json:"game_status"`  // "in_progress", "finished"
	Winner      string         `json:"winner"`       // "X", "O", "Draw", or ""
	PlayerXName string         `json:"player_x_name"`
	PlayerOName string         `json:"player_o_name"`
	UGNMoves    []string       `json:"ugn_moves"` // Array of UGN notation moves
	IsYourTurn  bool           `json:"is_your_turn"`
}

type BoardStateData struct {
	Boards      [9]SmallBoardData `json:"boards"`
	BoardStates [9]string         `json:"board_states"` // "undecided", "X", "O", "draw"
}

type SmallBoardData struct {
	Cells [9]string `json:"cells"` // "X", "O", or ""
	State string    `json:"state"` // "undecided", "X", "O", "draw"
}

type MovePayload struct {
	PlayerName   string `json:"player_name"`
	PlayerSymbol string `json:"player_symbol"`
	Move         string `json:"move"`        // e.g., "E5"
	BoardIndex   int    `json:"board_index"` // 0-8
	Position     int    `json:"position"`    // 0-8
}

type ErrorPayload struct {
	Message string `json:"message"`
}

type InfoPayload struct {
	Message string `json:"message"`
}

type GameOverPayload struct {
	Winner     string `json:"winner"`      // "X", "O", or "Draw"
	WinnerName string `json:"winner_name"` // Player name or "Draw"
	Message    string `json:"message"`
	Comment    string `json:"comment"` // e.g., "X wins by resignation"
}

type DrawOfferPayload struct {
	OfferedBy string `json:"offered_by"`
	Message   string `json:"message"`
}

func (ub *UltimateBoard) GetBoardStateData() BoardStateData {
	data := BoardStateData{}

	for i := 0; i < 9; i++ {
		data.Boards[i] = SmallBoardData{}
		for j := 0; j < 9; j++ {
			switch ub.Boards[i].Cells[j] {
			case X:
				data.Boards[i].Cells[j] = "X"
			case O:
				data.Boards[i].Cells[j] = "O"
			default:
				data.Boards[i].Cells[j] = ""
			}
		}

		switch ub.Boards[i].State {
		case XWins:
			data.Boards[i].State = "X"
			data.BoardStates[i] = "X"
		case OWins:
			data.Boards[i].State = "O"
			data.BoardStates[i] = "O"
		case Draw:
			data.Boards[i].State = "draw"
			data.BoardStates[i] = "draw"
		default:
			data.Boards[i].State = "undecided"
			data.BoardStates[i] = "undecided"
		}
	}

	return data
}

func (gs *GameSession) GetUGNMoves() []string {
	gs.mutex.RLock()
	defer gs.mutex.RUnlock()

	if gs.Logger == nil {
		return []string{}
	}

	movesStr := gs.Logger.GetUGNMovesString()
	if movesStr == "" {
		return []string{}
	}

	return splitMoves(movesStr)
}

func splitMoves(movesStr string) []string {
	if movesStr == "" {
		return []string{}
	}

	moves := make([]string, 0)
	current := ""

	for _, char := range movesStr {
		if char == ' ' {
			if current != "" {
				moves = append(moves, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}

	if current != "" {
		moves = append(moves, current)
	}

	return moves
}
