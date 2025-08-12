# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=boilerplate-api
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the application
.PHONY: build
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/api

# Install dependencies
.PHONY: deps
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Install development tools
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Development tools installed successfully!"
	@echo "Note: gosec is included in golangci-lint, no separate installation needed."

# Run tests
.PHONY: test
test:
	$(GOTEST) -v -coverprofile=coverage.out ./...

# Run tests with short flag
.PHONY: test-short
test-short:
	$(GOTEST) -short -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage: test
	$(GOCMD) tool cover -html=coverage.out

# Clean build files
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

# Run the application
.PHONY: run
run:
	$(GOCMD) run ./cmd/api

# Build for Linux
.PHONY: build-linux
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/api

# Format code
.PHONY: fmt
fmt:
	$(GOCMD) fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	@golangci-lint --version > /dev/null 2>&1 || (echo "golangci-lint not found. Installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# Security check (using golangci-lint with gosec enabled)
.PHONY: security
security:
	@echo "Running security checks via golangci-lint..."
	@golangci-lint --version > /dev/null 2>&1 || (echo "golangci-lint not found. Installing..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run --enable gosec

# Generate swagger docs
.PHONY: docs
docs:
	@echo "Generating Swagger documentation..."
	@swag -version > /dev/null 2>&1 || (echo "swag not found. Installing..." && go install github.com/swaggo/swag/cmd/swag@latest)
	swag init -g cmd/api/main.go

# Database migration up
.PHONY: migrate-up
migrate-up:
	migrate -path migrations -database "postgres://postgres:password@localhost:5432/boilerplate?sslmode=disable" -verbose up

# Database migration down
.PHONY: migrate-down
migrate-down:
	migrate -path migrations -database "postgres://postgres:password@localhost:5432/boilerplate?sslmode=disable" -verbose down

# Docker build
.PHONY: docker-build
docker-build:
	docker build -t $(BINARY_NAME) .

# Docker run
.PHONY: docker-run
docker-run:
	docker run -p 8080:8080 $(BINARY_NAME)

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build         Build the application"
	@echo "  deps          Install dependencies"
	@echo "  test          Run tests"
	@echo "  test-coverage Run tests with coverage"
	@echo "  clean         Clean build files"
	@echo "  run           Run the application"
	@echo "  build-linux   Build for Linux"
	@echo "  fmt           Format code"
	@echo "  lint          Lint code"
	@echo "  security      Run security checks"
	@echo "  docs          Generate swagger docs"
	@echo "  migrate-up    Run database migrations up"
	@echo "  migrate-down  Run database migrations down"
	@echo "  docker-build  Build Docker image"
	@echo "  docker-run    Run Docker container"
	@echo "  help          Show this help message"
	@echo "  install-tools Install development tools (golangci-lint, gosec, swag)"