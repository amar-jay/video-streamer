# Video Streamer Makefile

# Variables
BINARY_NAME := video-streamer
SERVICE_NAME := video-streamer
SERVICE_FILE := $(SERVICE_NAME).service
SYSTEMD_DIR := /etc/systemd/system
GO_FILES := $(wildcard *.go) $(wildcard internal/**/*.go)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

# Default target
.PHONY: all
all: build

# Build the application
.PHONY: build
build: $(BINARY_NAME)

$(BINARY_NAME): $(GO_FILES)
	@echo "🔨 Building $(BINARY_NAME)..."
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) server.go
	@echo "✅ Build complete: $(BINARY_NAME)"

# Build with debug information
.PHONY: build-debug
build-debug:
	@echo "🔨 Building $(BINARY_NAME) with debug info..."
	go build -gcflags="all=-N -l" -o $(BINARY_NAME) server.go
	@echo "✅ Debug build complete: $(BINARY_NAME)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	go clean
	@echo "✅ Clean complete"

# Run tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	go test ./...
	@echo "✅ Tests complete"

# Run the application locally (for development)
.PHONY: run
run: build
	@echo "🚀 Starting $(BINARY_NAME) locally..."
	./$(BINARY_NAME)

# Run with custom parameters
.PHONY: run-custom
run-custom: build
	@echo "🚀 Starting $(BINARY_NAME) with custom settings..."
	./$(BINARY_NAME) --input $(INPUT) --rtsp-address $(RTSP_ADDR)

# Format Go code
.PHONY: fmt
fmt:
	@echo "📝 Formatting Go code..."
	go fmt ./...
	@echo "✅ Formatting complete"

# Lint Go code
.PHONY: lint
lint:
	@echo "🔍 Linting Go code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not found, using go vet instead"; \
		go vet ./...; \
	fi
	@echo "✅ Linting complete"

# Download dependencies
.PHONY: deps
deps:
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "✅ Dependencies updated"

# Install the systemd service
.PHONY: install-service
install-service: build
	@echo "🔧 Installing systemd service..."
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "❌ Please run 'sudo make install-service' to install the service"; \
		exit 1; \
	fi
	@if [ ! -f "$(SERVICE_FILE)" ]; then \
		echo "❌ Service file '$(SERVICE_FILE)' not found"; \
		exit 1; \
	fi
	systemctl stop $(SERVICE_NAME) 2>/dev/null || true
	cp $(SERVICE_FILE) $(SYSTEMD_DIR)/
	systemctl daemon-reload
	systemctl enable $(SERVICE_NAME)
	@echo "✅ Service installed successfully"
	@echo "   Start with: sudo systemctl start $(SERVICE_NAME)"

# Uninstall the systemd service
.PHONY: uninstall-service
uninstall-service:
	@echo "🗑️  Uninstalling systemd service..."
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "❌ Please run 'sudo make uninstall-service' to uninstall the service"; \
		exit 1; \
	fi
	systemctl stop $(SERVICE_NAME) 2>/dev/null || true
	systemctl disable $(SERVICE_NAME) 2>/dev/null || true
	rm -f $(SYSTEMD_DIR)/$(SERVICE_FILE)
	systemctl daemon-reload
	@echo "✅ Service uninstalled successfully"

# Start the service
.PHONY: start-service
start-service:
	@echo "▶️  Starting $(SERVICE_NAME) service..."
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "❌ Please run 'sudo make start-service' to start the service"; \
		exit 1; \
	fi
	systemctl start $(SERVICE_NAME)
	@echo "✅ Service started"

# Stop the service
.PHONY: stop-service
stop-service:
	@echo "⏹️  Stopping $(SERVICE_NAME) service..."
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "❌ Please run 'sudo make stop-service' to stop the service"; \
		exit 1; \
	fi
	systemctl stop $(SERVICE_NAME)
	@echo "✅ Service stopped"

# Restart the service
.PHONY: restart-service
restart-service:
	@echo "🔄 Restarting $(SERVICE_NAME) service..."
	@if [ "$$(id -u)" -ne 0 ]; then \
		echo "❌ Please run 'sudo make restart-service' to restart the service"; \
		exit 1; \
	fi
	systemctl restart $(SERVICE_NAME)
	@echo "✅ Service restarted"

# Check service status
.PHONY: status-service
status-service:
	@echo "📊 Checking $(SERVICE_NAME) service status..."
	systemctl status $(SERVICE_NAME)

# View service logs
.PHONY: logs
logs:
	@echo "📋 Viewing $(SERVICE_NAME) service logs..."
	journalctl -u $(SERVICE_NAME) -f

# View recent service logs
.PHONY: logs-recent
logs-recent:
	@echo "📋 Viewing recent $(SERVICE_NAME) service logs..."
	journalctl -u $(SERVICE_NAME) --since "1 hour ago"

# Full deployment (build + install + start)
.PHONY: deploy
deploy: build install-service start-service
	@echo "🚀 Deployment complete!"
	@echo "   Service is running at rtsp://localhost:8554/"

# Development setup
.PHONY: dev-setup
dev-setup: deps fmt lint test build
	@echo "🛠️  Development setup complete"

# Check if service file syntax is valid
.PHONY: validate-service
validate-service:
	@echo "✅ Validating service file..."
	@if [ ! -f "$(SERVICE_FILE)" ]; then \
		echo "❌ Service file '$(SERVICE_FILE)' not found"; \
		exit 1; \
	fi
	systemd-analyze verify $(SERVICE_FILE) 2>/dev/null || echo "⚠️  systemd-analyze not available, skipping validation"
	@echo "✅ Service file validation complete"

# Show help
.PHONY: help
help:
	@echo "📖 Video Streamer Makefile Help"
	@echo ""
	@echo "🔨 Build Commands:"
	@echo "  make build          - Build the application"
	@echo "  make build-debug    - Build with debug information"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "🧪 Development Commands:"
	@echo "  make test           - Run tests"
	@echo "  make fmt            - Format Go code"
	@echo "  make lint           - Lint Go code"
	@echo "  make deps           - Download/update dependencies"
	@echo "  make dev-setup      - Complete development setup"
	@echo ""
	@echo "🚀 Run Commands:"
	@echo "  make run            - Run locally (development)"
	@echo "  make run-custom     - Run with custom params (INPUT=file RTSP_ADDR=:port)"
	@echo ""
	@echo "🔧 Service Commands (require sudo):"
	@echo "  make install-service   - Install systemd service"
	@echo "  make uninstall-service - Uninstall systemd service"
	@echo "  make start-service     - Start the service"
	@echo "  make stop-service      - Stop the service"
	@echo "  make restart-service   - Restart the service"
	@echo "  make status-service    - Check service status"
	@echo "  make deploy            - Full deployment (build + install + start)"
	@echo ""
	@echo "📋 Monitoring Commands:"
	@echo "  make logs           - Follow service logs"
	@echo "  make logs-recent    - View recent logs"
	@echo ""
	@echo "✅ Utility Commands:"
	@echo "  make validate-service - Validate service file"
	@echo "  make help             - Show this help"
	@echo ""
	@echo "📝 Examples:"
	@echo "  make build"
	@echo "  make run"
	@echo "  sudo make deploy"
	@echo "  make run-custom INPUT=/path/to/video.mp4 RTSP_ADDR=:9554"
	@echo ""
	@echo "📚 For detailed documentation, see README.md"

# Default help when no target is specified
.DEFAULT_GOAL := help
