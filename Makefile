# Video Streamer Makefile

BINARY_NAME := nebula-video-streamer
SERVICE_NAME := nebula-video-streamer
GO_FILES := $(wildcard *.go) $(wildcard internal/**/*.go)

# Default target
.PHONY: all build clean run install-service start stop status logs deploy help
all: build

# Build the application
build: $(BINARY_NAME)

$(BINARY_NAME): $(GO_FILES)
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) server.go

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	go clean

# Run locally
run: build
	./$(BINARY_NAME)

# Install systemd service
install-service: build
	@if [ "$$(id -u)" -ne 0 ]; then echo "Run with sudo"; exit 1; fi
	systemctl stop nebula-video-streamer 2>/dev/null || true
	cp nebula-video-streamer.service /etc/systemd/system/
	systemctl daemon-reload
	systemctl enable nebula-video-streamer
	@echo "Service installed. Start with: sudo systemctl start nebula-video-streamer"

# Start service
start:
	@if [ "$$(id -u)" -ne 0 ]; then echo "Run with sudo"; exit 1; fi
	systemctl start nebula-video-streamer

# Stop service
stop:
	@if [ "$$(id -u)" -ne 0 ]; then echo "Run with sudo"; exit 1; fi
	systemctl stop nebula-video-streamer

# Check service status
status:
	systemctl status nebula-video-streamer

# View logs
logs:
	journalctl -u nebula-video-streamer -f

# Deploy (build + install + start)
deploy: build install-service start
	@echo "Deployed! Access at rtsp://localhost:8554/"

# Show help
help:
	@echo "Video Streamer Makefile"
	@echo ""
	@echo "Commands:"
	@echo "  make build    - Build the application"
	@echo "  make run      - Run locally"
	@echo "  make clean    - Clean build files"
	@echo ""
	@echo "Service (requires sudo):"
	@echo "  make install-service - Install systemd service"
	@echo "  make start    - Start service"
	@echo "  make stop     - Stop service"
	@echo "  make status   - Check service status"
	@echo "  make logs     - View service logs"
	@echo "  make deploy   - Full deployment"

.DEFAULT_GOAL := help
