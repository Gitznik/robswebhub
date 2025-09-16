package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/gitznik/robswebhub/internal/database"
	"github.com/gitznik/robswebhub/internal/testhelpers"
	"github.com/google/uuid"
)

func TestDatabase_CreateAndGetMatch(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer db.Cleanup()

	ctx := context.Background()
	matchID := uuid.New()

	// Create match
	match, err := db.Queries.CreateMatch(ctx, database.CreateMatchParams{
		ID:        matchID,
		Player1:   "TestPlayer1",
		Player2:   "TestPlayer2",
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}

	// Verify creation
	if match.ID != matchID {
		t.Errorf("Expected match ID %v, got %v", matchID, match.ID)
	}
	if match.Player1 != "TestPlayer1" {
		t.Errorf("Expected Player1 to be TestPlayer1, got %s", match.Player1)
	}
	if match.Player2 != "TestPlayer2" {
		t.Errorf("Expected Player2 to be TestPlayer2, got %s", match.Player2)
	}

	// Get match
	retrievedMatch, err := db.Queries.GetMatch(ctx, matchID)
	if err != nil {
		t.Fatalf("Failed to get match: %v", err)
	}

	// Verify retrieval
	if retrievedMatch.ID != matchID {
		t.Errorf("Expected retrieved match ID %v, got %v", matchID, retrievedMatch.ID)
	}
	if retrievedMatch.Player1 != "TestPlayer1" {
		t.Errorf("Expected retrieved Player1 to be TestPlayer1, got %s", retrievedMatch.Player1)
	}
}

func TestDatabase_CreateAndGetScores(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer db.Cleanup()

	ctx := context.Background()

	// Create match first
	matchID := uuid.New()
	_, err := db.Queries.CreateMatch(ctx, database.CreateMatchParams{
		ID:        matchID,
		Player1:   "Alice",
		Player2:   "Bob",
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}

	// Create multiple scores
	gameIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	playDates := []time.Time{
		time.Now().AddDate(0, 0, -3),
		time.Now().AddDate(0, 0, -2),
		time.Now().AddDate(0, 0, -1),
	}

	for i, gameID := range gameIDs {
		winner := "Alice"
		if i%2 == 0 {
			winner = "Bob"
		}

		err := db.Queries.CreateScore(ctx, database.CreateScoreParams{
			MatchID:     matchID,
			GameID:      gameID,
			Winner:      winner,
			WinnerScore: int16(2 + i),
			LoserScore:  1,
			CreatedAt:   time.Now(),
			PlayedAt:    playDates[i],
		})
		if err != nil {
			t.Fatalf("Failed to create score %d: %v", i, err)
		}
	}

	// Get match scores
	cutoffDate := time.Now().AddDate(0, -1, 0)
	scores, err := db.Queries.GetMatchScores(ctx, database.GetMatchScoresParams{
		MatchID:  matchID,
		PlayedAt: cutoffDate,
	})
	if err != nil {
		t.Fatalf("Failed to get match scores: %v", err)
	}

	// Verify scores
	if len(scores) != 3 {
		t.Errorf("Expected 3 scores, got %d", len(scores))
	}

	// Verify scores are ordered by played_at DESC
	for i := 0; i < len(scores)-1; i++ {
		if scores[i].PlayedAt.Before(scores[i+1].PlayedAt) {
			t.Errorf("Scores not ordered correctly: %v before %v", scores[i].PlayedAt, scores[i+1].PlayedAt)
		}
	}

	// Get recent scores
	recentScores, err := db.Queries.GetRecentScores(ctx, matchID)
	if err != nil {
		t.Fatalf("Failed to get recent scores: %v", err)
	}

	// Verify recent scores (should return max 5)
	if len(recentScores) != 3 {
		t.Errorf("Expected 3 recent scores, got %d", len(recentScores))
	}
}

func TestDatabase_BulkCreateScores(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer db.Cleanup()

	ctx := context.Background()

	// Create match
	matchID := uuid.New()
	_, err := db.Queries.CreateMatch(ctx, database.CreateMatchParams{
		ID:        matchID,
		Player1:   "Player1",
		Player2:   "Player2",
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create match: %v", err)
	}

	// Prepare bulk scores
	bulkScores := []database.CreateBulkScoresParams{}
	for i := 0; i < 10; i++ {
		winner := "Player1"
		if i%2 == 0 {
			winner = "Player2"
		}

		bulkScores = append(bulkScores, database.CreateBulkScoresParams{
			MatchID:     matchID,
			GameID:      uuid.New(),
			Winner:      winner,
			WinnerScore: int16(3),
			LoserScore:  int16(2),
			CreatedAt:   time.Now(),
			PlayedAt:    time.Now().AddDate(0, 0, -i),
		})
	}

	// Bulk create
	rowsAffected, err := db.Queries.CreateBulkScores(ctx, bulkScores)
	if err != nil {
		t.Fatalf("Failed to bulk create scores: %v", err)
	}

	if rowsAffected != 10 {
		t.Errorf("Expected 10 rows affected, got %d", rowsAffected)
	}

	// Verify all scores were created
	scores, err := db.Queries.GetMatchScores(ctx, database.GetMatchScoresParams{
		MatchID:  matchID,
		PlayedAt: time.Now().AddDate(0, -1, 0),
	})
	if err != nil {
		t.Fatalf("Failed to get scores: %v", err)
	}

	if len(scores) != 10 {
		t.Errorf("Expected 10 scores, got %d", len(scores))
	}
}

func TestDatabase_EdgeCases(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer db.Cleanup()

	ctx := context.Background()

	t.Run("Get non-existent match", func(t *testing.T) {
		_, err := db.Queries.GetMatch(ctx, uuid.New())
		if err == nil {
			t.Error("Expected error when getting non-existent match")
		}
	})

	t.Run("Get scores for non-existent match", func(t *testing.T) {
		scores, err := db.Queries.GetMatchScores(ctx, database.GetMatchScoresParams{
			MatchID:  uuid.New(),
			PlayedAt: time.Now().AddDate(0, -1, 0),
		})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if len(scores) != 0 {
			t.Errorf("Expected 0 scores, got %d", len(scores))
		}
	})

	t.Run("Create score with same game_id", func(t *testing.T) {
		matchID := uuid.New()
		gameID := uuid.New()

		// Create match
		_, err := db.Queries.CreateMatch(ctx, database.CreateMatchParams{
			ID:        matchID,
			Player1:   "P1",
			Player2:   "P2",
			CreatedAt: time.Now(),
		})
		if err != nil {
			t.Fatalf("Failed to create match: %v", err)
		}

		// Create first score
		err = db.Queries.CreateScore(ctx, database.CreateScoreParams{
			MatchID:     matchID,
			GameID:      gameID,
			Winner:      "P1",
			WinnerScore: 2,
			LoserScore:  1,
			CreatedAt:   time.Now(),
			PlayedAt:    time.Now(),
		})
		if err != nil {
			t.Fatalf("Failed to create first score: %v", err)
		}

		// Try to create duplicate
		err = db.Queries.CreateScore(ctx, database.CreateScoreParams{
			MatchID:     matchID,
			GameID:      gameID,
			Winner:      "P2",
			WinnerScore: 3,
			LoserScore:  1,
			CreatedAt:   time.Now(),
			PlayedAt:    time.Now(),
		})
		if err == nil {
			t.Error("Expected error when creating duplicate game_id")
		}
	})
}

func TestDatabase_Transactions(t *testing.T) {
	db := testhelpers.SetupTestDB(t)
	defer db.Cleanup()

	ctx := context.Background()

	t.Run("Transaction rollback on error", func(t *testing.T) {
		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}
		defer func() {
			if err := tx.Rollback(ctx); err != nil {
				t.Logf("Failed to rollback transaction: %v", err)
			}
		}()

		qtx := db.Queries.WithTx(tx)

		// Create match in transaction
		matchID := uuid.New()
		_, err = qtx.CreateMatch(ctx, database.CreateMatchParams{
			ID:        matchID,
			Player1:   "TxPlayer1",
			Player2:   "TxPlayer2",
			CreatedAt: time.Now(),
		})
		if err != nil {
			t.Fatalf("Failed to create match in transaction: %v", err)
		}

		// Rollback
		if err := tx.Rollback(ctx); err != nil {
			t.Logf("Failed to rollback transaction: %v", err)
		}

		// Verify match doesn't exist
		_, err = db.Queries.GetMatch(ctx, matchID)
		if err == nil {
			t.Error("Expected match not to exist after rollback")
		}
	})

	t.Run("Transaction commit", func(t *testing.T) {
		tx, err := db.Pool.Begin(ctx)
		if err != nil {
			t.Fatalf("Failed to begin transaction: %v", err)
		}

		qtx := db.Queries.WithTx(tx)

		// Create match in transaction
		matchID := uuid.New()
		_, err = qtx.CreateMatch(ctx, database.CreateMatchParams{
			ID:        matchID,
			Player1:   "CommitPlayer1",
			Player2:   "CommitPlayer2",
			CreatedAt: time.Now(),
		})
		if err != nil {
			t.Fatalf("Failed to create match in transaction: %v", err)
		}

		// Commit
		err = tx.Commit(ctx)
		if err != nil {
			t.Fatalf("Failed to commit transaction: %v", err)
		}

		// Verify match exists
		match, err := db.Queries.GetMatch(ctx, matchID)
		if err != nil {
			t.Errorf("Expected match to exist after commit: %v", err)
		}
		if match.Player1 != "CommitPlayer1" {
			t.Errorf("Expected Player1 to be CommitPlayer1, got %s", match.Player1)
		}
	})
}
