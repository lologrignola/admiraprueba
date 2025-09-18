.PHONY: build run test clean docker-build docker-run help

# Default target
all: build

# Build the application
build:
	@echo "Building admira-etl..."
	go build -o bin/admira-etl main.go

# Run the application
run: build
	@echo "Starting admira-etl..."
	./bin/admira-etl

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t admira-etl .

# Docker run
docker-run: docker-build
	@echo "Running Docker container..."
	docker-compose up

# Docker run in background
docker-run-bg: docker-build
	@echo "Running Docker container in background..."
	docker-compose up -d

# Stop Docker containers
docker-stop:
	@echo "Stopping Docker containers..."
	docker-compose down

# View Docker logs
docker-logs:
	@echo "Viewing Docker logs..."
	docker-compose logs -f

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the application"
	@echo "  run            - Build and run the application"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Build and run with Docker Compose"
	@echo "  docker-run-bg  - Run Docker containers in background"
	@echo "  docker-stop    - Stop Docker containers"
	@echo "  docker-logs    - View Docker logs"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  deps           - Install dependencies"
	@echo "  help           - Show this help message"

