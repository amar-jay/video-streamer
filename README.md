# Video Streamer

Go-based RTSP video streaming server that streams video files over RTSP protocol.

## ✨ Features

- 🎥 Stream video files via RTSP protocol
- 🔧 CLI interface with customizable parameters
- 🚀 systemd service support for production deployment
- �️ Security hardened service configuration
- 📊 Comprehensive logging and monitoring
- 🔄 Automatic restart on failure

## 🚀 Quick Start

### Development Mode
```bash
# Build and run locally
make build
make run
```

### Production Deployment
```bash
# Deploy as systemd service
sudo make deploy

# Monitor service
make logs
make status
```

## 📦 Installation

### Prerequisites
- Go 1.24.2 or later
- Linux system with systemd (for service deployment)
- TLS certificate files (`server.crt` and `server.key`) for secure connections

### Generate TLS Certificates (if needed)
```bash
# Generate private key
openssl genrsa -out server.key 2048

# Generate self-signed certificate
openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
```

### Build from Source
```bash
git clone <repository-url>
cd video-streamer
make build
```

## 🔧 Configuration

### Command Line Options
```bash
./video-streamer [OPTIONS]

Options:
  --input, -i     Path to input video file (default: "/home/amarjay/Downloads/demo.mp4")
  --rtsp-address  RTSP server address (default: ":8554")
  --udp-rtp-address   UDP RTP address (default: ":8000")
  --udp-rtcp-address  UDP RTCP address (default: ":8001")
  --help, -h      Show help
```

### Examples
```bash
# Basic usage
./video-streamer --input /path/to/video.mp4

# Custom ports
./video-streamer --input /path/to/video.mp4 --rtsp-address :9554

# Bind to specific interface
./video-streamer --rtsp-address 192.168.1.100:8554
```

## 🎯 Makefile Commands

### Development
```bash
make build          # Build the application
make clean          # Clean build artifacts
make run            # Run locally
```

### Service Management (requires sudo)
```bash
make install-service # Install systemd service
make start          # Start service
make stop           # Stop service
make status         # Check service status
make logs           # View service logs
make deploy         # Build + install + start
```

## 🔧 systemd Service

The application can be deployed as a systemd service for production use.

### Service Installation
```bash
# Build the application
make build

# Install and start service
sudo make deploy

# Check service status
sudo systemctl status video-streamer
```

### Service Management
```bash
# Control service
sudo systemctl start video-streamer
sudo systemctl stop video-streamer
sudo systemctl restart video-streamer

# Enable/disable auto-start
sudo systemctl enable video-streamer
sudo systemctl disable video-streamer

# View logs
sudo journalctl -u video-streamer -f
```

### Service Configuration
The service is installed to `/etc/systemd/system/video-streamer.service`.

To customize configuration:
1. Edit the service file: `sudo nano /etc/systemd/system/video-streamer.service`
2. Modify the `ExecStart` line with your desired parameters
3. Reload and restart: `sudo systemctl daemon-reload && sudo systemctl restart video-streamer`

Example custom configuration:
```ini
ExecStart=/home/amarjay/Desktop/code/video-streamer/video-streamer \
  --input /media/videos/stream.mp4 \
  --rtsp-address :9554
```

### Security Features
The systemd service includes security hardening:
- Runs as non-root user
- Protected system directories
- Private temporary directory
- Resource limits
- No new privileges

## 📡 Accessing the Stream

Once running, access your RTSP stream at:

### Local Access
```
rtsp://localhost:8554/
```

### Remote Access
```
rtsp://YOUR_SERVER_IP:8554/
```

### Test with Media Players
```bash
# VLC
vlc rtsp://localhost:8554/

# FFplay
ffplay rtsp://localhost:8554/

# GStreamer
gst-launch-1.0 rtspsrc location=rtsp://localhost:8554/ ! decodebin ! autovideosink
```

## 🐛 Troubleshooting

### Common Issues

**Service won't start:**
1. Check if binary exists: `ls -la video-streamer`
2. Verify video file permissions: `ls -la /path/to/video.mp4`
3. View logs: `sudo journalctl -u video-streamer -f`

**Permission issues:**
- Ensure user has read access to video file
- Check working directory permissions

**Network issues:**
- Verify ports aren't in use: `sudo netstat -tlnp | grep :8554`
- Check firewall settings for remote access

**Video file issues:**
- Ensure video file is readable and valid
- Check supported formats (H.264 recommended)

### Debug Mode
```bash
# Build with debug symbols
make build-debug

# Run with verbose logging
./video-streamer --input /path/to/video.mp4 --debug
```

## 🗂️ Project Structure

```
.
├── server.go                  # Main application
├── internal/
│   ├── server/
│   │   └── handler.go         # RTSP server handler
│   ├── streamer/
│   │   └── streamer.go        # File streaming logic
│   └── utils/
│       └── video_utils.go     # Video utilities
├── video-streamer.service     # systemd service file
├── install-service.sh         # Service installation script
├── uninstall-service.sh       # Service removal script
├── Makefile                   # Build automation
└── README.md                  # This file
```

## � License

[Add your license information here]

## 🤝 Contributing

[Add contributing guidelines here]
# 1. Setup
mkdir video-streamer && cd video-streamer
go mod init video-streamer

# 2. Install GStreamer
sudo apt-get install gstreamer1.0-tools gstreamer1.0-plugins-*

# 3. Build & Run
make build
./build/video-streamer

# 4. Test
vlc rtsp://localhost:8554/stream
```

## 🎯 Architecture

```
Video Source → GStreamer (H.264) → RTP → RTSP Server → Clients
```

**Estimated Time: 20-30 hours**
**Key Libraries: gortsplib, Pion RTP, GStreamer CLI**
