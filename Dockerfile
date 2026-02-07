# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make curl

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Install templ and sqlc
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Download Tailwind CSS standalone CLI
ARG TAILWIND_VERSION=v4.1.18
RUN ARCH=$(uname -m) && \
  case "$ARCH" in x86_64) ARCH="x64" ;; aarch64) ARCH="arm64" ;; esac && \
  curl -sL -o /usr/local/bin/tailwindcss \
  "https://github.com/tailwindlabs/tailwindcss/releases/download/${TAILWIND_VERSION}/tailwindcss-linux-${ARCH}" && \
  chmod +x /usr/local/bin/tailwindcss

# Copy source code
COPY . .

# Generate code
RUN templ generate
RUN sqlc generate
RUN tailwindcss -i static/css/input.css -o static/css/styles.css --minify

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/server .

# Copy static files and config
COPY --from=builder /app/static ./static
COPY --from=builder /app/config ./config
COPY --from=builder /app/migrations ./migrations

# Create directory for chart images
RUN mkdir -p ./static/images/match_plots

EXPOSE 8080

CMD ["./server"]
