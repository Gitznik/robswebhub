# RobsWebHub

My personal website built with Go, featuring personal projects and an introduction about myself.

## 🚀 Quick Start

```bash
# Clone the repository
git clone https://github.com/gitznik/robswebhub.git
cd robswebhub

# Run the setup script (installs everything and starts the database)
just setup

# Start the development server
just dev

# Visit http://localhost:8000
```

## Tech Stack

- **Backend**: Go with Gin web framework
- **Database**: PostgreSQL with sqlc for type-safe queries
- **Migrations**: golang-migrate
- **Templates**: templ for type-safe HTML templates
- **Frontend**: HTMX for dynamic interactions, Pico CSS for styling
- **Charts**: go-echarts for data visualization
- **Build Tool**: [just](https://github.com/casey/just) for task automation

## Why just?

This project uses `just` instead of `make` for task automation because:
- **Better syntax**: Cleaner, more readable command definitions
- **Built-in help**: Automatic help text generation
- **Cross-platform**: Works consistently across Linux, macOS, and Windows
- **Shell flexibility**: Choose your shell per recipe
- **Better error messages**: More helpful when things go wrong
- **Parameters**: Easy command parameters without make's complexity

## Features

- Personal portfolio and about page
- **Scorekeeper**: A game score tracking application
  - Single and batch score entry
  - Visual charts showing win progression
  - Historical score tracking

## Prerequisites

- Go 1.22 or later
- PostgreSQL
- Docker (optional, for containerized development)
- [just](https://github.com/casey/just) (command runner, install with `cargo install just` or `brew install just`)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/gitznik/robswebhub.git
cd robswebhub
```

2. Install just (if not already installed):
```bash
# macOS
brew install just

# Linux/WSL
curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to /usr/local/bin

# Or with Cargo
cargo install just
```

3. Install dependencies and tools:
```bash
just install
```

Or manually:
```bash
go mod download
go install github.com/a-h/templ/cmd/templ@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

4. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

## Development

### Quick Start with Docker Compose

```bash
# Start PostgreSQL and the application
docker-compose up -d

# Stop everything
docker-compose down
```

### Manual Setup

1. Start PostgreSQL:
```bash
just setup-env
# Or use your existing PostgreSQL instance
```

2. Run migrations:
```bash
just migrate-up
```

3. Generate code (templ templates and sqlc queries):
```bash
just generate
```

4. Run the application:
```bash
just run
```

For hot reload during development (requires [air](https://github.com/cosmtrek/air)):
```bash
go install github.com/cosmtrek/air@latest
just dev
```

### Available Commands

```bash
just --list          # Show all available commands
just setup           # Complete initial setup
just start           # Start everything (DB, migrations, seed, dev server)
just stop            # Stop all services
just dev             # Run with hot reload
just test            # Run tests
just check           # Run all checks (fmt, lint, test)
just db-reset        # Reset database completely
just compose-up      # Start with docker-compose
just backup-db       # Backup database
just help            # Show detailed help
```

### Advanced Commands

The justfile includes many advanced commands for development:

```bash
# Database Management
just db-console              # Connect to database with psql
just migrate-status          # Check migration status
just migrate-goto <version>  # Go to specific migration
just backup-db               # Create timestamped backup
just restore-db <file>       # Restore from backup

# Development Tools
just watch                   # Watch files and auto-rebuild
just profile-cpu            # Profile CPU usage
just profile-mem            # Profile memory usage
just outdated               # Check for outdated dependencies
just audit                  # Run security audit

# Docker & Deployment
just docker                 # Build and run Docker container
just compose-logs          # View docker-compose logs
just deploy                # Deploy to Fly.io
just fly-ssh               # SSH into Fly.io app
```

For a complete list of commands with descriptions:
```bash
just help
```

## Project Structure

```
robswebhub/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection and generated sqlc code
│   ├── handlers/        # HTTP request handlers
│   └── templates/       # Templ templates
│       ├── layouts/     # Base layouts
│       ├── pages/       # Page templates
│       └── components/  # Reusable components
├── migrations/          # Database migrations
├── static/              # Static assets (CSS, JS, images)
├── config/              # Configuration files
└── scripts/             # Utility scripts
```

## Configuration

The application uses a hierarchical configuration system:

1. Base configuration from `config/dev.yaml` or `config/production.yaml`
2. Environment variables (prefixed with `APP_`)
3. `.env` file (for local development)

### Environment Variables

- `APP_ENVIRONMENT`: `dev` or `production`
- `APP_APPLICATION__HOST`: Server host
- `APP_APPLICATION__PORT`: Server port
- `DATABASE_URL`: PostgreSQL connection string

## Database Schema

### Matches Table
- `id`: UUID (Primary Key)
- `player_1`: Text
- `player_2`: Text
- `created_at`: Timestamp

### Scores Table
- `match_id`: UUID (Foreign Key)
- `game_id`: UUID
- `winner`: Text
- `winner_score`: SmallInt
- `loser_score`: SmallInt
- `played_at`: Date
- `created_at`: Timestamp

## Deployment

### Using Docker

```bash
# Build the image
just docker-build

# Run the container
just docker-run

# Or do both
just docker
```

### Fly.io Deployment

```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh

# Deploy
fly deploy
```

## API Routes

- `GET /` - Home page
- `GET /about` - About page
- `GET /scores` - Scorekeeper index
- `POST /scores/single` - Submit single score
- `POST /scores/batch` - Submit batch scores
- `GET /scores/single-form` - Get single score form (HTMX)
- `GET /scores/batch-form` - Get batch score form (HTMX)
- `GET /scores/chart/:id` - Get match chart

## Development Tips

1. **Code Generation**: Always run `just generate` after modifying `.templ` files or `queries.sql`

2. **Database Changes**: Create new migrations with:
   ```bash
   just migrate-create your_migration_name
   ```

3. **Testing**: The example match ID `b13a16d8-c46e-4921-83f2-eec9675fce74` is seeded for testing

4. **Hot Reload**: Use `just dev` for automatic reloading during development

5. **Quick Commands**: 
   - `just start` - Start everything (DB, migrations, seed, dev server)
   - `just quick` - Quickly rebuild and run after changes
   - `just check` - Run all pre-commit checks (fmt, lint, test)

## Contributing

Feel free to open issues or submit pull requests if you find any bugs or have suggestions for improvements.

## Author

**Robert Offner**
- GitHub: [@Gitznik](https://github.com/Gitznik)
- LinkedIn: [Robert Offner](https://www.linkedin.com/in/robert-offner-065993191)

## License

This project is open source. Feel free to use it as a template for your own projects.