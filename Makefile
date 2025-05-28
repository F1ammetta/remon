.PHONY: build run test clean docker-build docker-run dev install-air templ-generate

# Generate templ files
templ-generate:
	@which templ > /dev/null || (echo "Installing templ..." && go install github.com/a-h/templ/cmd/templ@latest)
	templ generate

# Build the application
build: templ-generate
	go build -o bin/system-monitor .

# Run the application
run: templ-generate
	go run .

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/ tmp/
	find . -name "*_templ.go" -delete

# Build Docker image
docker-build:
	docker build -t system-monitor .

# Run with Docker Compose
docker-run:
	docker-compose up -d

# Stop Docker Compose
docker-stop:
	docker-compose down

# View logs
logs:
	docker-compose logs -f

# Install dependencies
deps:
	go mod tidy
	go mod download

# Install Air for live reloading
install-air:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)

# Install templ
install-templ:
	@which templ > /dev/null || (echo "Installing templ..." && go install github.com/a-h/templ/cmd/templ@latest)

# Development server with auto-reload using Air
dev: install-air install-templ
	@echo "Starting development server with live reload..."
	@mkdir -p tmp
	@mkdir -p static
	@cp styles.css static/ 2>/dev/null || true
	@cp script.js static/ 2>/dev/null || true
	air

# Development server with verbose output
dev-verbose: install-air install-templ
	@echo "Starting development server with live reload (verbose)..."
	@mkdir -p tmp
	@mkdir -p static
	@cp styles.css static/ 2>/dev/null || true
	@cp script.js static/ 2>/dev/null || true
	air -d

# Format code
fmt:
	go fmt ./...
	templ fmt .

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Install development tools
install-dev-tools:
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Setup development environment
setup-dev: install-dev-tools deps
	@echo "Development environment setup complete!"
	@echo "Run 'make dev' to start the development server with live reload"

# Production build
build-prod: templ-generate
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/system-monitor .

# Run production build
run-prod: build-prod
	./bin/system-monitor

# Watch templ files for changes
watch-templ:
	templ generate --watch

