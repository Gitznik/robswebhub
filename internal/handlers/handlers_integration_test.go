package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gitznik/robswebhub/internal/models"
	"github.com/gitznik/robswebhub/internal/testhelpers"
	"github.com/google/uuid"
)

func TestIntegration_HomePage(t *testing.T) {
	t.Parallel()

	// Test GET /
	t.Run("GET / returns home page", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
		testhelpers.AssertResponseContains(t, w.Body.String(), "RobsWebHub")
		testhelpers.AssertResponseContains(t, w.Body.String(), "Scorekeeper")
	})

	// Test HEAD /
	t.Run("HEAD / returns 200", func(t *testing.T) {
		req, _ := http.NewRequest("HEAD", "/", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
		if w.Body.Len() > 0 {
			t.Errorf("HEAD request should not return body, got %d bytes", w.Body.Len())
		}
	})
}

func TestIntegration_AboutPage(t *testing.T) {
	t.Parallel()

	t.Run("GET /about returns about page", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/about", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
		testhelpers.AssertResponseContains(t, w.Body.String(), "About me")
		testhelpers.AssertResponseContains(t, w.Body.String(), "Robert Offner")
	})
}

func TestIntegration_ScoresIndex(t *testing.T) {
	t.Parallel()

	testData := db.SeedTestData(t)

	t.Run("GET /scores without matchup_id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/scores", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
		testhelpers.AssertResponseContains(t, w.Body.String(), "Scorekeeper")
		testhelpers.AssertResponseContains(t, w.Body.String(), "example match")
	})

	t.Run("GET /scores with valid matchup_id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/scores?matchup_id=%s", testData.MatchID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
		testhelpers.AssertResponseContains(t, w.Body.String(), testData.MatchID.String())
		testhelpers.AssertResponseContains(t, w.Body.String(), "Add Score")
		testhelpers.AssertResponseContains(t, w.Body.String(), "Most recent results")
	})

	t.Run("GET /scores with invalid matchup_id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/scores?matchup_id=invalid-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
		testhelpers.AssertResponseContains(t, w.Body.String(), "Invalid matchup ID")
	})

	t.Run("GET /scores with non-existent matchup_id", func(t *testing.T) {
		nonExistentID := uuid.New()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/scores?matchup_id=%s", nonExistentID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
		testhelpers.AssertResponseContains(t, w.Body.String(), "Match not found")
	})
}

func TestIntegration_SingleScoreSubmission(t *testing.T) {
	t.Parallel()
	// Setup
	testData := db.SeedTestData(t)

	t.Run("POST /scores/single with valid data", func(t *testing.T) {
		formData := url.Values{
			"matchup_id":      {testData.MatchID.String()},
			"winner_initials": {testData.Player1},
			"score":           {"3:1"},
			"played_at":       {time.Now().Format("2006-01-02")},
		}

		req, _ := http.NewRequest("POST", "/scores/single", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should redirect on success
		testhelpers.AssertResponseCode(t, http.StatusSeeOther, w.Code)

		location := w.Header().Get("Location")
		if !strings.Contains(location, fmt.Sprintf("/scores?matchup_id=%s", testData.MatchID)) {
			t.Errorf("Expected redirect to contain matchup_id, got: %s", location)
		}
	})

	t.Run("POST /scores/single with invalid player", func(t *testing.T) {
		formData := url.Values{
			"matchup_id":      {testData.MatchID.String()},
			"winner_initials": {"INVALID"},
			"score":           {"3:1"},
			"played_at":       {time.Now().Format("2006-01-02")},
		}

		req, _ := http.NewRequest("POST", "/scores/single", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusSeeOther, w.Code)

		location := w.Header().Get("Location")
		if !strings.Contains(location, "Player not in match") {
			t.Errorf("Expected error about player not in match, got: %s", location)
		}
	})

	t.Run("POST /scores/single with invalid score format", func(t *testing.T) {
		formData := url.Values{
			"matchup_id":      {testData.MatchID.String()},
			"winner_initials": {testData.Player1},
			"score":           {"invalid"},
			"played_at":       {time.Now().Format("2006-01-02")},
		}

		req, _ := http.NewRequest("POST", "/scores/single", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusSeeOther, w.Code)

		location := w.Header().Get("Location")
		if !strings.Contains(location, "Invalid score format") {
			t.Errorf("Expected error about invalid score format, got: %s", location)
		}
	})
}

func TestIntegration_BatchScoreSubmission(t *testing.T) {
	t.Parallel()
	// Setup
	testData := db.SeedTestData(t)

	t.Run("POST /scores/batch with valid data", func(t *testing.T) {
		batchData := fmt.Sprintf(`%s %s 2:1
%s %s 3:2
%s %s 2:0`,
			time.Now().AddDate(0, 0, -3).Format("2006-01-02"), testData.Player1,
			time.Now().AddDate(0, 0, -2).Format("2006-01-02"), testData.Player2,
			time.Now().AddDate(0, 0, -1).Format("2006-01-02"), testData.Player1,
		)

		formData := url.Values{
			"matchup_id":       {testData.MatchID.String()},
			"raw_matches_list": {batchData},
		}

		req, _ := http.NewRequest("POST", "/scores/batch", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusSeeOther, w.Code)

		location := w.Header().Get("Location")
		if !strings.Contains(location, fmt.Sprintf("/scores?matchup_id=%s", testData.MatchID)) {
			t.Errorf("Expected redirect to contain matchup_id, got: %s", location)
		}
	})

	t.Run("POST /scores/batch with mixed valid and invalid data", func(t *testing.T) {
		batchData := fmt.Sprintf(`%s %s 2:1
invalid line
%s %s 3:2`,
			time.Now().AddDate(0, 0, -3).Format("2006-01-02"), testData.Player1,
			time.Now().AddDate(0, 0, -1).Format("2006-01-02"), testData.Player2,
		)

		formData := url.Values{
			"matchup_id":       {testData.MatchID.String()},
			"raw_matches_list": {batchData},
		}

		req, _ := http.NewRequest("POST", "/scores/batch", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should still redirect, skipping invalid lines
		testhelpers.AssertResponseCode(t, http.StatusSeeOther, w.Code)
	})
}

func TestIntegration_ScoreForms(t *testing.T) {
	t.Parallel()
	// Setup
	testData := db.SeedTestData(t)

	t.Run("GET /scores/single-form", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/scores/single-form?matchup_id=%s", testData.MatchID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
		testhelpers.AssertResponseContains(t, w.Body.String(), "Single Entry")
		testhelpers.AssertResponseContains(t, w.Body.String(), "winner_initials")
	})

	t.Run("GET /scores/batch-form", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/scores/batch-form?matchup_id=%s", testData.MatchID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
		testhelpers.AssertResponseContains(t, w.Body.String(), "Batch Entry")
		testhelpers.AssertResponseContains(t, w.Body.String(), "raw_matches_list")
	})
}

func TestIntegration_CloudPage(t *testing.T) {
	t.Parallel()

	t.Run("GET /cloud returns cloud page with all services", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/cloud", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
		testhelpers.AssertResponseContains(t, w.Body.String(), "RobsCloud")

		for _, svc := range models.CloudServices() {
			testhelpers.AssertResponseContains(t, w.Body.String(), svc.Name)
			testhelpers.AssertResponseContains(t, w.Body.String(), svc.Description)
			testhelpers.AssertResponseContains(t, w.Body.String(), svc.URL)
		}
	})
}

func TestIntegration_ScoresChart(t *testing.T) {
	t.Parallel()
	// Setup
	testData := db.SeedTestData(t)

	t.Run("GET /scores/chart/:id with valid match", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/scores/chart/%s", testData.MatchID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusOK, w.Code)
		// Chart response should contain chart HTML/JS
		testhelpers.AssertResponseContains(t, w.Body.String(), "Summary of Wins")
	})

	t.Run("GET /scores/chart/:id with invalid UUID", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/scores/chart/invalid-uuid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusBadRequest, w.Code)
		testhelpers.AssertResponseContains(t, w.Body.String(), "Invalid match ID")
	})

	t.Run("GET /scores/chart/:id with non-existent match", func(t *testing.T) {
		nonExistentID := uuid.New()
		req, _ := http.NewRequest("GET", fmt.Sprintf("/scores/chart/%s", nonExistentID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		testhelpers.AssertResponseCode(t, http.StatusNotFound, w.Code)
		testhelpers.AssertResponseContains(t, w.Body.String(), "Match not found")
	})
}
