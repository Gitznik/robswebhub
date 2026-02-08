package handlers_test

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/database"
	"github.com/gitznik/robswebhub/internal/testhelpers"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
)

var (
	db     *testhelpers.TestDB
	router *gin.Engine
)

func TestMain(m *testing.M) {
	db = testhelpers.SetupTestDB()

	router = testhelpers.SetupTestRouter(db.Queries)
	m.Run()

	db.Pool.Close()
	if err := testcontainers.TerminateContainer(db.Container); err != nil {
		log.Printf("Failed to terminate container: %v", err)
	}
}

// TestE2E_CompleteScoreWorkflow tests the complete scorekeeper workflow
func TestE2E_CompleteScoreWorkflow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Step 1: Create a new match
	t.Run("Step 1: Create new match", func(t *testing.T) {
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

		// Step 2: View the empty scorecard
		t.Run("Step 2: View empty scorecard", func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/scores?matchup_id=%s", matchID), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
			testhelpers.AssertResponseContains(t, w.Body.String(), matchID.String())
			testhelpers.AssertResponseContains(t, w.Body.String(), "Add Score")
		})

		// Step 3: Add single score
		t.Run("Step 3: Add single score", func(t *testing.T) {
			formData := url.Values{
				"matchup_id":      {matchID.String()},
				"winner_initials": {"Alice"},
				"score":           {"3:1"},
				"played_at":       {time.Now().Format("2006-01-02")},
			}

			req, _ := http.NewRequest("POST", "/scores/single", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			testhelpers.AssertResponseCode(t, http.StatusSeeOther, w.Code)
		})

		// Step 4: Add batch scores
		t.Run("Step 4: Add batch scores", func(t *testing.T) {
			batchData := fmt.Sprintf(`%s Alice 2:1
%s Bob 3:2
%s Alice 2:0
%s Bob 3:1
%s Alice 2:1`,
				time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
				time.Now().AddDate(0, 0, -4).Format("2006-01-02"),
				time.Now().AddDate(0, 0, -3).Format("2006-01-02"),
				time.Now().AddDate(0, 0, -2).Format("2006-01-02"),
				time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			)

			formData := url.Values{
				"matchup_id":       {matchID.String()},
				"raw_matches_list": {batchData},
			}

			req, _ := http.NewRequest("POST", "/scores/batch", strings.NewReader(formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			testhelpers.AssertResponseCode(t, http.StatusSeeOther, w.Code)
		})

		// Step 5: View updated scorecard with results
		t.Run("Step 5: View scorecard with results", func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/scores?matchup_id=%s", matchID), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
			testhelpers.AssertResponseContains(t, w.Body.String(), "Most recent results")
			testhelpers.AssertResponseContains(t, w.Body.String(), "Alice")
			testhelpers.AssertResponseContains(t, w.Body.String(), "Bob")
		})

		// Step 6: View chart
		t.Run("Step 6: View wins chart", func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/scores/chart/%s", matchID), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
			testhelpers.AssertResponseContains(t, w.Body.String(), "Summary of Wins")
			testhelpers.AssertResponseContains(t, w.Body.String(), "Wins of Alice")
			testhelpers.AssertResponseContains(t, w.Body.String(), "Wins of Bob")
		})

		// Step 7: Verify data persistence
		t.Run("Step 7: Verify data persistence", func(t *testing.T) {
			// Check match exists in database
			match, err := db.Queries.GetMatch(ctx, matchID)
			if err != nil {
				t.Fatalf("Failed to retrieve match: %v", err)
			}
			if match.Player1 != "Alice" || match.Player2 != "Bob" {
				t.Errorf("Expected players Alice and Bob, got %s and %s", match.Player1, match.Player2)
			}

			// Check scores exist
			cutoffDate := time.Now().AddDate(0, -6, 0)
			scores, err := db.Queries.GetMatchScores(ctx, database.GetMatchScoresParams{
				MatchID:  matchID,
				PlayedAt: cutoffDate,
			})
			if err != nil {
				t.Fatalf("Failed to retrieve scores: %v", err)
			}
			if len(scores) != 6 { // 1 single + 5 batch
				t.Errorf("Expected 6 scores, got %d", len(scores))
			}

			// Count wins
			aliceWins := 0
			bobWins := 0
			for _, score := range scores {
				switch score.Winner {
				case "Alice":
					aliceWins++
				default:
					bobWins++
				}
			}
			if aliceWins != 4 || bobWins != 2 {
				t.Errorf("Expected 4 Alice wins and 2 Bob wins, got %d and %d", aliceWins, bobWins)
			}
		})
	})
}

// TestE2E_MultipleMatchesWorkflow tests managing multiple matches
func TestE2E_MultipleMatchesWorkflow(t *testing.T) {
	// Setup
	ctx := context.Background()

	// Create multiple matches
	matches := []struct {
		ID      uuid.UUID
		Player1 string
		Player2 string
	}{
		{uuid.New(), "John", "Jane"},
		{uuid.New(), "Alice", "Bob"},
		{uuid.New(), "Charlie", "Diana"},
	}

	for _, match := range matches {
		_, err := db.Queries.CreateMatch(ctx, database.CreateMatchParams{
			ID:        match.ID,
			Player1:   match.Player1,
			Player2:   match.Player2,
			CreatedAt: time.Now(),
		})
		if err != nil {
			t.Fatalf("Failed to create match: %v", err)
		}
	}

	// Add scores to each match
	for i, match := range matches {
		t.Run(fmt.Sprintf("Add scores to match %d", i+1), func(t *testing.T) {
			// Add different number of scores to each match
			for j := 0; j <= i+2; j++ {
				winner := match.Player1
				if j%2 == 0 {
					winner = match.Player2
				}

				err := db.Queries.CreateScore(ctx, database.CreateScoreParams{
					MatchID:     match.ID,
					GameID:      uuid.New(),
					Winner:      winner,
					WinnerScore: int16(2 + j),
					LoserScore:  int16(1),
					CreatedAt:   time.Now(),
					PlayedAt:    time.Now().AddDate(0, 0, -j),
				})
				if err != nil {
					t.Fatalf("Failed to create score: %v", err)
				}
			}
		})
	}

	// Navigate between matches
	for i, match := range matches {
		t.Run(fmt.Sprintf("View match %d scorecard", i+1), func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/scores?matchup_id=%s", match.ID), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
			testhelpers.AssertResponseContains(t, w.Body.String(), match.Player1)
			testhelpers.AssertResponseContains(t, w.Body.String(), match.Player2)
		})
	}
}

// TestE2E_ErrorHandlingWorkflow tests error scenarios
func TestE2E_ErrorHandlingWorkflow(t *testing.T) {
	t.Run("Handle non-existent match gracefully", func(t *testing.T) {
		nonExistentID := uuid.New()

		// Try to view non-existent match
		req, _ := http.NewRequest("GET", fmt.Sprintf("/scores?matchup_id=%s", nonExistentID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
		testhelpers.AssertResponseContains(t, w.Body.String(), "Match not found")

		// Try to add score to non-existent match
		formData := url.Values{
			"matchup_id":      {nonExistentID.String()},
			"winner_initials": {"Player1"},
			"score":           {"2:1"},
			"played_at":       {time.Now().Format("2006-01-02")},
		}

		req, _ = http.NewRequest("POST", "/scores/single", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusSeeOther, w.Code)
		location := w.Header().Get("Location")
		if !strings.Contains(location, "error") {
			t.Errorf("Expected error in redirect, got: %s", location)
		}
	})

	t.Run("Handle malformed input gracefully", func(t *testing.T) {
		// Create a valid match first
		matchID := uuid.New()
		ctx := context.Background()
		_, err := db.Queries.CreateMatch(ctx, database.CreateMatchParams{
			ID:        matchID,
			Player1:   "Player1",
			Player2:   "Player2",
			CreatedAt: time.Now(),
		})
		if err != nil {
			t.Fatalf("Failed to create match: %v", err)
		}

		// Test various malformed inputs
		malformedInputs := []struct {
			name     string
			formData url.Values
			expected string
		}{
			{
				name: "Empty score",
				formData: url.Values{
					"matchup_id":      {matchID.String()},
					"winner_initials": {"Player1"},
					"score":           {""},
					"played_at":       {time.Now().Format("2006-01-02")},
				},
				expected: "error",
			},
			{
				name: "Invalid date format",
				formData: url.Values{
					"matchup_id":      {matchID.String()},
					"winner_initials": {"Player1"},
					"score":           {"2:1"},
					"played_at":       {"not-a-date"},
				},
				expected: "error",
			},
			{
				name: "Negative scores",
				formData: url.Values{
					"matchup_id":      {matchID.String()},
					"winner_initials": {"Player1"},
					"score":           {"-1:2"},
					"played_at":       {time.Now().Format("2006-01-02")},
				},
				expected: "error",
			},
		}

		for _, tc := range malformedInputs {
			t.Run(tc.name, func(t *testing.T) {
				req, _ := http.NewRequest("POST", "/scores/single", strings.NewReader(tc.formData.Encode()))
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				testhelpers.AssertResponseCode(t, http.StatusSeeOther, w.Code)
				location := w.Header().Get("Location")
				if !strings.Contains(location, tc.expected) {
					t.Errorf("Expected %s in redirect, got: %s", tc.expected, location)
				}
			})
		}
	})
}

// TestE2E_NavigationWorkflow tests navigation between pages
func TestE2E_NavigationWorkflow(t *testing.T) {
	navigationPaths := []struct {
		name     string
		path     string
		expected string
	}{
		{"Home", "/", "RobsWebHub"},
		{"About", "/about", "Robert Offner"},
		{"Scores", "/scores", "Scorekeeper"},
		{"Back to Home from Scores", "/", "Projects"},
	}

	for _, nav := range navigationPaths {
		t.Run(fmt.Sprintf("Navigate to %s", nav.name), func(t *testing.T) {
			req, _ := http.NewRequest("GET", nav.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
			testhelpers.AssertResponseContains(t, w.Body.String(), nav.expected)
		})
	}
}
