package testhelpers

import (
	"context"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gitznik/robswebhub/internal/config"
	"github.com/gitznik/robswebhub/internal/database"
	"github.com/gitznik/robswebhub/internal/handlers"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDB holds test database connection and cleanup function
type TestDB struct {
	Pool      *pgxpool.Pool
	Queries   *database.Queries
	Container testcontainers.Container
	ConnStr   string
	cleanup   func()
}

// SetupTestDB creates a test database using testcontainers
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	ctx := context.Background()

	// Start PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Connect to database
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	m, err := migrate.New(
		GetMigrationsPath(),
		connStr,
	)
	if err != nil {
		t.Fatalf("Failed to create migration instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("Failed to run migrations: %v", err)
	}
	m.Close()

	queries := database.New(pool)

	return &TestDB{
		Pool:      pool,
		Queries:   queries,
		Container: postgresContainer,
		ConnStr:   connStr,
		cleanup: func() {
			pool.Close()
			if err := postgresContainer.Terminate(ctx); err != nil {
				t.Logf("Failed to terminate container: %v", err)
			}
		},
	}
}

// Cleanup cleans up test database resources
func (tdb *TestDB) Cleanup() {
	if tdb.cleanup != nil {
		tdb.cleanup()
	}
}

// SeedTestData seeds the database with test data
func (tdb *TestDB) SeedTestData(t *testing.T) TestData {
	t.Helper()

	ctx := context.Background()
	testData := TestData{
		MatchID:  uuid.New(),
		Player1:  "P1",
		Player2:  "P2",
		GameIDs:  []uuid.UUID{},
		PlayedAt: time.Now().AddDate(0, 0, -7),
	}

	// Create a test match
	_, err := tdb.Queries.CreateMatch(ctx, database.CreateMatchParams{
		ID:        testData.MatchID,
		Player1:   testData.Player1,
		Player2:   testData.Player2,
		CreatedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("Failed to create test match: %v", err)
	}

	// Create some test scores
	for i := 0; i < 5; i++ {
		gameID := uuid.New()
		testData.GameIDs = append(testData.GameIDs, gameID)

		winner := testData.Player1
		if i%2 == 0 {
			winner = testData.Player2
		}

		err := tdb.Queries.CreateScore(ctx, database.CreateScoreParams{
			MatchID:     testData.MatchID,
			GameID:      gameID,
			Winner:      winner,
			WinnerScore: 2,
			LoserScore:  1,
			CreatedAt:   time.Now(),
			PlayedAt:    testData.PlayedAt.AddDate(0, 0, i),
		})
		if err != nil {
			t.Fatalf("Failed to create test score: %v", err)
		}
	}

	return testData
}

// TestData holds test data references
type TestData struct {
	MatchID  uuid.UUID
	Player1  string
	Player2  string
	GameIDs  []uuid.UUID
	PlayedAt time.Time
}

// SetupTestRouter creates a test Gin router
func SetupTestRouter(queries *database.Queries) *gin.Engine {
	gin.SetMode(gin.TestMode)
	return setupRouter(&config.Config{
		Application: config.ApplicationConfig{
			Environment: "test",
		},
	}, queries)
}

// This should match the setupRouter function from main.go
func setupRouter(cfg *config.Config, queries *database.Queries) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	// Serve static files
	router.Static("/static", "./static")

	h := handlers.New(queries)

	// Routes
	router.GET("/", h.Home)
	router.HEAD("/", h.HomeHead)
	router.GET("/about", h.About)

	// Scores routes
	scores := router.Group("/scores")
	{
		scores.GET("", h.ScoresIndex)
		scores.POST("/single", h.ScoresSingle)
		scores.POST("/batch", h.ScoresBatch)
		scores.GET("/single-form", h.SingleScoreForm)
		scores.GET("/batch-form", h.BatchScoreForm)
		scores.GET("/chart/:id", h.ScoresChart)
	}

	return router
}

// AssertResponseCode checks if response code matches expected
func AssertResponseCode(t *testing.T, expected, actual int) {
	t.Helper()
	if expected != actual {
		t.Errorf("Expected response code %d, got %d", expected, actual)
	}
}

// AssertResponseContains checks if response body contains expected string
func AssertResponseContains(t *testing.T, body, expected string) {
	t.Helper()
	if !contains(body, expected) {
		t.Errorf("Expected response to contain '%s', but it didn't. Body: %s", expected, body)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 &&
			(s[:len(substr)] == substr || contains(s[1:], substr))))
}

// CreateTestConfig creates a test configuration
func CreateTestConfig(dbConnStr string) *config.Config {
	return &config.Config{
		Application: config.ApplicationConfig{
			Host:        "127.0.0.1",
			Port:        8080,
			Environment: "test",
		},
		Database: config.DatabaseConfig{
			ConnectionString: dbConnStr,
		},
	}
}
