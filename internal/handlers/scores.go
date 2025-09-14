package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/database"
	"github.com/gitznik/robswebhub/internal/templates/components"
	"github.com/gitznik/robswebhub/internal/templates/pages"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/google/uuid"
)

type ScoreInput struct {
	MatchupID      uuid.UUID `form:"matchup_id" binding:"required"`
	WinnerInitials string    `form:"winner_initials" binding:"required"`
	Score          string    `form:"score" binding:"required"`
	PlayedAt       string    `form:"played_at" binding:"required"`
}

type BatchScoreInput struct {
	MatchupID      uuid.UUID `form:"matchup_id" binding:"required"`
	RawMatchesList string    `form:"raw_matches_list" binding:"required"`
}

func (h *Handler) ScoresIndex(c *gin.Context) {
	matchupIDStr := c.Query("matchup_id")

	var matchupID uuid.UUID
	var match *database.GetMatchRow
	var scores []database.GetMatchScoresRow
	var recentScores []database.GetRecentScoresRow
	var err error

	if matchupIDStr != "" {
		matchupID, err = uuid.Parse(matchupIDStr)
		if err != nil {
			component := pages.Scores(nil, nil, nil, "Invalid matchup ID")
			component.Render(c.Request.Context(), c.Writer)
			return
		}

		// Get match information
		result, err := h.queries.GetMatch(c.Request.Context(), matchupID)
		if err != nil {
			component := pages.Scores(nil, nil, nil, "Match not found")
			component.Render(c.Request.Context(), c.Writer)
			return
		}
		match = &result

		// Get match scores (last 6 months)
		cutoffDate := time.Now().AddDate(0, -6, 0)
		scores, err = h.queries.GetMatchScores(c.Request.Context(), database.GetMatchScoresParams{
			MatchID:  matchupID,
			PlayedAt: cutoffDate,
		})
		if err != nil {
			scores = []database.GetMatchScoresRow{}
		}

		// Get recent scores for display
		recentScores, err = h.queries.GetRecentScores(c.Request.Context(), matchupID)
		if err != nil {
			recentScores = []database.GetRecentScoresRow{}
		}
	}

	component := pages.Scores(match, scores, recentScores, "")
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		c.String(http.StatusInternalServerError, "Failed to render page")
		return
	}
}

func (h *Handler) ScoresSingle(c *gin.Context) {
	var input ScoreInput
	if err := c.ShouldBind(&input); err != nil {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/scores?matchup_id=%s&error=%s", input.MatchupID, err.Error()))
		return
	}

	// Validate match exists and player is in match
	match, err := h.queries.GetMatch(c.Request.Context(), input.MatchupID)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/scores?error=Match not found")
		return
	}

	if input.WinnerInitials != match.Player1 && input.WinnerInitials != match.Player2 {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/scores?matchup_id=%s&error=Player not in match", input.MatchupID))
		return
	}

	// Parse score
	scoreParts := strings.Split(input.Score, ":")
	if len(scoreParts) != 2 {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/scores?matchup_id=%s&error=Invalid score format", input.MatchupID))
		return
	}

	winnerScore, err := strconv.ParseInt(scoreParts[0], 10, 16)
	if err != nil {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/scores?matchup_id=%s&error=Invalid winner score", input.MatchupID))
		return
	}

	loserScore, err := strconv.ParseInt(scoreParts[1], 10, 16)
	if err != nil {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/scores?matchup_id=%s&error=Invalid loser score", input.MatchupID))
		return
	}

	// Make sure winner score is higher
	if winnerScore <= loserScore {
		winnerScore, loserScore = loserScore, winnerScore
	}

	// Parse date
	playedAt, err := time.Parse("2006-01-02", input.PlayedAt)
	if err != nil {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/scores?matchup_id=%s&error=Invalid date format", input.MatchupID))
		return
	}

	// Save score
	err = h.queries.CreateScore(c.Request.Context(), database.CreateScoreParams{
		MatchID:     input.MatchupID,
		GameID:      uuid.New(),
		Winner:      input.WinnerInitials,
		WinnerScore: int16(winnerScore),
		LoserScore:  int16(loserScore),
		CreatedAt:   time.Now(),
		PlayedAt:    playedAt,
	})

	if err != nil {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/scores?matchup_id=%s&error=%s", input.MatchupID, err.Error()))
		return
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/scores?matchup_id=%s", input.MatchupID))
}

func (h *Handler) ScoresBatch(c *gin.Context) {
	var input BatchScoreInput
	if err := c.ShouldBind(&input); err != nil {
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/scores?matchup_id=%s&error=%s", input.MatchupID, err.Error()))
		return
	}

	// Validate match exists
	match, err := h.queries.GetMatch(c.Request.Context(), input.MatchupID)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/scores?error=Match not found")
		return
	}

	// Parse batch input
	lines := strings.Split(strings.ReplaceAll(input.RawMatchesList, "\r\n", "\n"), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) != 3 {
			continue // Skip invalid lines
		}

		date := parts[0]
		winner := parts[1]
		score := parts[2]

		// Validate winner is in match
		if winner != match.Player1 && winner != match.Player2 {
			continue
		}

		// Parse date
		playedAt, err := time.Parse("2006-01-02", date)
		if err != nil {
			continue
		}

		// Parse score
		scoreParts := strings.Split(score, ":")
		if len(scoreParts) != 2 {
			continue
		}

		winnerScore, err := strconv.ParseInt(scoreParts[0], 10, 16)
		if err != nil {
			continue
		}

		loserScore, err := strconv.ParseInt(scoreParts[1], 10, 16)
		if err != nil {
			continue
		}

		// Make sure winner score is higher
		if winnerScore <= loserScore {
			winnerScore, loserScore = loserScore, winnerScore
		}

		// Save score
		h.queries.CreateScore(context.Background(), database.CreateScoreParams{
			MatchID:     input.MatchupID,
			GameID:      uuid.New(),
			Winner:      winner,
			WinnerScore: int16(winnerScore),
			LoserScore:  int16(loserScore),
			CreatedAt:   time.Now(),
			PlayedAt:    playedAt,
		})
	}

	c.Redirect(http.StatusSeeOther, fmt.Sprintf("/scores?matchup_id=%s", input.MatchupID))
}

func (h *Handler) SingleScoreForm(c *gin.Context) {
	matchupIDStr := c.Query("matchup_id")
	var matchupID *uuid.UUID

	if matchupIDStr != "" {
		id, err := uuid.Parse(matchupIDStr)
		if err == nil {
			matchupID = &id
		}
	}

	component := components.SingleScoreForm(matchupID)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		c.String(http.StatusInternalServerError, "Failed to render form")
		return
	}
}

func (h *Handler) BatchScoreForm(c *gin.Context) {
	matchupIDStr := c.Query("matchup_id")
	var matchupID *uuid.UUID

	if matchupIDStr != "" {
		id, err := uuid.Parse(matchupIDStr)
		if err == nil {
			matchupID = &id
		}
	}

	component := components.BatchScoreForm(matchupID)
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		c.String(http.StatusInternalServerError, "Failed to render form")
		return
	}
}

func (h *Handler) ScoresChart(c *gin.Context) {
	matchIDStr := c.Param("id")
	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid match ID")
		return
	}

	// Get match information
	match, err := h.queries.GetMatch(c.Request.Context(), matchID)
	if err != nil {
		c.String(http.StatusNotFound, "Match not found")
		return
	}

	// Get scores
	cutoffDate := time.Now().AddDate(0, -6, 0)
	scores, err := h.queries.GetMatchScores(c.Request.Context(), database.GetMatchScoresParams{
		MatchID:  matchID,
		PlayedAt: cutoffDate,
	})
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to get scores")
		return
	}

	// Calculate cumulative wins
	p1Wins := []opts.LineData{}
	p2Wins := []opts.LineData{}
	p1Count := 0
	p2Count := 0

	// Group and sort by date
	dateMap := make(map[string][]database.GetMatchScoresRow)
	for _, score := range scores {
		dateKey := score.PlayedAt.Format("2006-01-02")
		dateMap[dateKey] = append(dateMap[dateKey], score)
	}

	// Process in chronological order
	dates := []string{}
	for date := range dateMap {
		dates = append(dates, date)
	}

	for _, date := range dates {
		for _, score := range dateMap[date] {
			if score.Winner == match.Player1 {
				p1Count++
			} else {
				p2Count++
			}
		}
		p1Wins = append(p1Wins, opts.LineData{Value: p1Count})
		p2Wins = append(p2Wins, opts.LineData{Value: p2Count})
	}

	// Create line chart
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme:  types.ThemeWesteros,
			Width:  "640px",
			Height: "480px",
		}),
		charts.WithTitleOpts(opts.Title{
			Title: "Summary of Wins",
		}),
	)

	line.SetXAxis(dates).
		AddSeries(fmt.Sprintf("Wins of %s", match.Player1), p1Wins).
		AddSeries(fmt.Sprintf("Wins of %s", match.Player2), p2Wins).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))

	line.Render(c.Writer)
}
