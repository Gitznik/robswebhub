package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/database"
	"github.com/gitznik/robswebhub/internal/middleware"
	"github.com/gitznik/robswebhub/internal/sessions"
	"github.com/gitznik/robswebhub/internal/templates/pages"
	"github.com/jackc/pgx/v5"
)

func (h *Handler) GamesIndex(c *gin.Context) {
	p, err := sessions.GetProfile(c)
	if err != nil {
		_ = c.Error(errors.New("Failed to read user information"))
		log.Printf("%v", err)
		return
	}
	playerID := p.Sub

	isSignedUp := true
	_, err = h.queries.GetPlayerInformation(c.Request.Context(), playerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			isSignedUp = false
		} else {
			_ = c.Error(errors.New("Failed to read players"))
			log.Printf("Failed to read players: %v", err)
			return
		}
	}

	if !isSignedUp {
		c.Redirect(http.StatusSeeOther, "/gamekeeper/signup")
		return
	}

	gamesOfUser, err := h.queries.ListGamesOfUser(c.Request.Context(), playerID)
	if err != nil {
		_ = c.Error(errors.New("Failed to read players games"))
		log.Printf("Failed to read players games: %v", err)
		return
	}

	component := pages.Games(gamesOfUser, "", c.GetBool(middleware.LoginKey))
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		_ = c.Error(errors.New("Failed to render page"))
		return
	}
}

func (h *Handler) SignUp(c *gin.Context) {
	p, err := sessions.GetProfile(c)
	if err != nil {
		_ = c.Error(errors.New("Failed to read user information"))
		log.Printf("%v", err)
		return
	}
	playerID := p.Sub

	isSignedUp := true
	_, err = h.queries.GetPlayerInformation(c.Request.Context(), playerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Printf("error is no rows")
			isSignedUp = false
		} else {
			_ = c.Error(errors.New("Failed to read players"))
			log.Printf("Failed to read players: %v", err)
			return
		}
	}

	if isSignedUp {
		c.Redirect(http.StatusSeeOther, "/gamekeeper")
		return
	}

	component := pages.GamesSignup("", c.GetBool(middleware.LoginKey))
	if err := component.Render(c.Request.Context(), c.Writer); err != nil {
		_ = c.Error(errors.New("Failed to render page"))
		return
	}
}

func (h *Handler) DoSignUp(c *gin.Context) {
	p, err := sessions.GetProfile(c)
	if err != nil {
		_ = c.Error(errors.New("Failed to read user information"))
		log.Printf("%v", err)
		return
	}
	playerID := p.Sub

	_, err = h.queries.CreatePlayer(c.Request.Context(), database.CreatePlayerParams{
		PlayerID:  playerID,
		CreatedAt: time.Now(),
	})
	if err != nil {
		_ = c.Error(errors.New("Could not sign up"))
		log.Printf("Failed to create player: %v", err)
		return
	}
	c.Header("HX-Redirect", "/gamekeeper")
	c.String(http.StatusCreated, "Signup sucessful")
}
