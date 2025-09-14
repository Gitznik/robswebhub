#!/bin/bash
set -e

echo "🚀 Setting up RobsWebHub development environment..."
echo ""

# Check for required tools
echo "📦 Checking required tools..."

if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.22 or later."
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo "❌ Docker is not installed. Please install Docker."
    exit 1
fi

if ! command -v psql &> /dev/null; then
    echo "⚠️  Warning: psql is not installed. Some database commands may not work."
fi

echo "✅ Required tools are installed"
echo ""

# Install Go tools
echo "📥 Installing Go development tools..."
go install github.com/a-h/templ/cmd/templ@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/air-verse/air@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
echo "✅ Go tools installed"
echo ""

# Install Go dependencies
echo "📚 Installing Go dependencies..."
go mod download
echo "✅ Dependencies installed"
echo ""

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "📝 Creating .env file from template..."
    cp .env.example .env
    echo "✅ .env file created (please update with your settings)"
else
    echo "✅ .env file already exists"
fi
echo ""

# Setup database
echo "🐘 Setting up PostgreSQL database..."
if [ -z "${SKIP_DB_SETUP}" ]; then
    ./scripts/init_postgres.sh
else
    echo "⏭️  Skipping database setup (SKIP_DB_SETUP is set)"
fi
echo ""

# Generate code
echo "🔨 Generating code..."
templ generate
sqlc generate
echo "✅ Code generated"
echo ""

echo "🎉 Setup complete! You can now run:"
echo ""
echo "  just run    # Run the application"
echo "  just dev    # Run with hot reload"
echo "  just help   # See all available commands"
echo ""
echo "The application will be available at http://localhost:8000"
echo "Example match: http://localhost:8000/scores?matchup_id=b13a16d8-c46e-4921-83f2-eec9675fce74"
