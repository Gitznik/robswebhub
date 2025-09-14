package models

import (
	"time"

	"github.com/google/uuid"
)

// Match represents a match between two players
type Match struct {
	ID        uuid.UUID `json:"id"`
	Player1   string    `json:"player_1"`
	Player2   string    `json:"player_2"`
	CreatedAt time.Time `json:"created_at"`
}

// Score represents a single game score within a match
type Score struct {
	MatchID     uuid.UUID `json:"match_id"`
	GameID      uuid.UUID `json:"game_id"`
	Winner      string    `json:"winner"`
	WinnerScore int16     `json:"winner_score"`
	LoserScore  int16     `json:"loser_score"`
	PlayedAt    time.Time `json:"played_at"`
	CreatedAt   time.Time `json:"created_at"`
}

// MatchSummary contains aggregated match statistics
type MatchSummary struct {
	Match        *Match
	TotalGames   int
	Player1Wins  int
	Player2Wins  int
	RecentScores []Score
}

// PlayerStats contains statistics for a player
type PlayerStats struct {
	PlayerName  string
	TotalWins   int
	TotalLosses int
	WinRate     float64
	AvgScore    float64
}
