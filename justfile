# Default recipe to display help information
default:
    @just --list

# Initial project setup - run this first!
setup:
    chmod +x scripts/setup.sh scripts/init_postgres.sh
    ./scripts/setup.sh

# Setup development environment with PostgreSQL
setup-env:
    ./scripts/init_postgres.sh

# Stop and remove PostgreSQL container
teardown-env:
    #!/usr/bin/env bash
    RUNNING_POSTGRES_CONTAINER=$(docker ps --filter 'name=postgres' --format '{{{{.ID}}')
    if [ -n "$RUNNING_POSTGRES_CONTAINER" ]; then
        docker kill $RUNNING_POSTGRES_CONTAINER
    fi

# Install dependencies and development tools
install:
    go mod download
    go install github.com/a-h/templ/cmd/templ@latest
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    go install github.com/cosmtrek/air@latest

# Generate templ files
templ:
    templ generate

# Generate sqlc code
sqlc:
    sqlc generate

# Generate all code (templ and sqlc)
generate: templ sqlc

# Run database migrations
migrate-up:
    migrate -path migrations -database "${DATABASE_URL:-postgres://postgres:password@localhost:5432/robswebhub?sslmode=disable}" up

# Rollback database migrations
migrate-down:
    migrate -path migrations -database "${DATABASE_URL:-postgres://postgres:password@localhost:5432/robswebhub?sslmode=disable}" down 1

# Create a new migration (usage: just migrate-create add_user_table)
migrate-create name:
    migrate create -ext sql -dir migrations -seq {{name}}

# Build the application
build: generate
    go build -o bin/server cmd/server/main.go

# Run the application
run: generate
    go run cmd/server/main.go

# Run application with hot reload (requires air)
dev:
    air

# Watch for changes and rebuild
watch:
    @just generate
    watchexec -e go,templ,sql -- just run

# Install test dependencies
install-test-deps:
    go install gotest.tools/gotestsum@latest
    go install github.com/testcontainers/testcontainers-go@latest
    go get -t ./...

# Run all tests with gotestsum
test-all: generate
    @echo "Running all tests with gotestsum..."
    gotestsum --format testname -- -v -race -coverprofile=coverage.out ./...

# Run integration tests only
test-integration: generate
    @echo "Running integration tests..."
    gotestsum --format testname -- -v -race -tags=integration ./... -run Integration

# Run end-to-end tests only
test-e2e: generate
    @echo "Running end-to-end tests..."
    gotestsum --format testname -- -v -race -tags=e2e ./... -run E2E

# Run tests with watch mode
test-watch: generate
    gotestsum --format testname --watch -- -v ./...

# Run tests with coverage report
test-coverage: generate
    @echo "Running tests with coverage..."
    gotestsum --format testname -- -v -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated at coverage.html"

# Run tests in CI mode (with JUnit output)
test-ci: generate
    gotestsum --junitfile test-results.xml --format standard-verbose -- -v -race -coverprofile=coverage.out ./...

# Run specific test
test-run name: generate
    gotestsum --format testname -- -v -run {{name}} ./...

# Run tests with database container
test-db: generate
    @echo "Running tests with database..."
    TEST_DB_CONTAINER=true gotestsum --format testname -- -v -race ./...

# Run linter (requires golangci-lint)
lint:
    golangci-lint run

# Format code
fmt:
    go fmt ./...
    templ fmt .

# Check for code issues
check: fmt lint

# Clean build artifacts
clean:
    rm -rf bin/
    rm -f coverage.out coverage.html
    rm -rf tmp/

# Build Docker image
docker-build:
    docker build -t robswebhub .

# Run Docker container
docker-run:
    docker run -p 8080:8080 \
        --env APP_ENVIRONMENT=production \
        --env APP_APPLICATION__PORT=8080 \
        --env DATABASE_URL="postgres://postgres:password@host.docker.internal:5432/robswebhub?sslmode=disable" \
        robswebhub

# Build and run Docker container
docker: docker-build docker-run

# Start application with docker-compose
compose-up:
    docker compose up -d

# Stop application with docker-compose
compose-down:
    docker compose down

# View docker-compose logs
compose-logs:
    docker compose logs -f

# Restart docker-compose services
compose-restart: compose-down compose-up

# Seed database with example data
seed-db:
    @echo "Inserting example match..."
    @psql "${DATABASE_URL:-postgres://postgres:password@localhost:5432/robswebhub?sslmode=disable}" \
        -c "INSERT INTO matches(id, player_1, player_2, created_at) VALUES ('b13a16d8-c46e-4921-83f2-eec9675fce74', 'P1', 'P2', now()) ON CONFLICT DO NOTHING;"
    @psql "${DATABASE_URL:-postgres://postgres:password@localhost:5432/robswebhub?sslmode=disable}" \
        -c "INSERT INTO scores(match_id, game_id, winner, created_at, winner_score, loser_score, played_at) VALUES ('b13a16d8-c46e-4921-83f2-eec9675fce74', 'b13a16d8-c46e-4921-83f2-eec9675fce75', 'P1', now(), 2, 1, now()) ON CONFLICT DO NOTHING;"
    @echo "Database seeded successfully!"

# Reset database (drop, recreate, migrate, seed)
db-reset:
    @echo "Resetting database..."
    @psql "${DATABASE_URL:-postgres://postgres:password@localhost:5432/postgres?sslmode=disable}" \
        -c "DROP DATABASE IF EXISTS robswebhub;"
    @psql "${DATABASE_URL:-postgres://postgres:password@localhost:5432/postgres?sslmode=disable}" \
        -c "CREATE DATABASE robswebhub;"
    @just migrate-up
    @just seed-db

# Connect to database with psql
db-console:
    psql "${DATABASE_URL:-postgres://postgres:password@localhost:5432/robswebhub?sslmode=disable}"

# Show database migration status
migrate-status:
    migrate -path migrations -database "${DATABASE_URL:-postgres://postgres:password@localhost:5432/robswebhub?sslmode=disable}" version

# Run a specific migration version
migrate-goto version:
    migrate -path migrations -database "${DATABASE_URL:-postgres://postgres:password@localhost:5432/robswebhub?sslmode=disable}" goto {{version}}

# Install development tools
install-dev-tools:
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    go install golang.org/x/tools/cmd/goimports@latest
    go install mvdan.cc/gofumpt@latest
    go install github.com/segmentio/golines@latest

# Run all pre-commit checks
pre-commit: fmt lint
    @echo "All pre-commit checks passed!"

# Update Go dependencies
update-deps:
    go get -u ./...
    go mod tidy

# Verify Go dependencies
verify-deps:
    go mod verify

# Run the application in production mode
prod:
    APP_ENVIRONMENT=production just run

# Deploy to Fly.io
deploy:
    fly deploy

# SSH into Fly.io app
fly-ssh:
    fly ssh console

# View Fly.io logs
fly-logs:
    fly logs

# Open Fly.io app in browser
fly-open:
    fly open

# Show Fly.io app status
fly-status:
    fly status

# Create a full backup of the database
backup-db timestamp=`date +%Y%m%d_%H%M%S`:
    @echo "Creating database backup..."
    @mkdir -p backups
    @pg_dump "${DATABASE_URL:-postgres://postgres:password@localhost:5432/robswebhub?sslmode=disable}" > backups/backup_{{timestamp}}.sql
    @echo "Backup saved to backups/backup_{{timestamp}}.sql"

# Restore database from backup
restore-db backup_file:
    @echo "Restoring database from {{backup_file}}..."
    @psql "${DATABASE_URL:-postgres://postgres:password@localhost:5432/robswebhub?sslmode=disable}" < {{backup_file}}
    @echo "Database restored successfully!"

# Run security audit
audit:
    go list -json -deps ./... | nancy sleuth
    gosec ./...

# Generate SQL documentation
sql-docs:
    @echo "Generating SQL documentation..."
    @sqlc compile
    @echo "SQL documentation generated!"

# Start all development services
start: setup-env migrate-up seed-db dev
    @echo "Development environment is running!"

# Stop all development services
stop: teardown-env
    @echo "Development environment stopped!"

# Quickly rebuild and run after changes
quick: generate run

# Run benchmarks
bench:
    go test -bench=. -benchmem ./...

# Profile CPU usage
profile-cpu:
    go test -cpuprofile=cpu.prof -bench=. ./...
    go tool pprof cpu.prof

# Profile memory usage
profile-mem:
    go test -memprofile=mem.prof -bench=. ./...
    go tool pprof mem.prof

# Check for outdated dependencies
outdated:
    go list -u -m all

# Print current version
version:
    @git describe --tags --always --dirty

# Help - show all available commands with descriptions
help:
    @echo "RobsWebHub - Available Commands"
    @echo "================================"
    @just --list --unsorted
