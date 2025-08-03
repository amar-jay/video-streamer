# Video Streamer - Build TODO 🎯

Go-based RTSP video streaming server with H.264 encoding via GStreamer.

## 📋 Main TODO List

### 1. Project Setup
- [ ] Create Go module: `go mod init video-streamer`
- [ ] Add dependencies: `gortsplib/v4`, `pion/rtp`
- [ ] Create directories: `cmd/server/`, `internal/{config,rtsp,rtp,stream}/`
- [ ] Basic Makefile

### 2. Config (`internal/config/config.go`)
- [ ] Config struct with RTSP port, video params, GStreamer settings
- [ ] Command-line flags parsing
- [ ] SDP generation method
- [ ] GStreamer pipeline builder

### 3. RTP Handler (`internal/rtp/handler.go`)
- [ ] Pion RTP integration
- [ ] H.264 NAL unit packetization
- [ ] FU-A fragmentation for large packets
- [ ] Timestamp/sequence management

### 4. RTSP Server (`internal/rtsp/server.go`)
- [ ] gortsplib server setup
- [ ] Handler methods: OnDescribe, OnSetup, OnPlay
- [ ] Session management
- [ ] H.264 media format configuration

### 5. Stream Manager (`internal/stream/manager.go`)
- [ ] GStreamer process launcher
- [ ] UDP listener for RTP packets
- [ ] Forward packets to RTSP clients
- [ ] Multi-client support

### 6. Main App (`cmd/server/main.go`)
- [ ] CLI interface
- [ ] Component initialization
- [ ] Graceful shutdown
- [ ] Error handling

### 7. Testing
- [ ] Build with `make build`
- [ ] Test with VLC: `vlc rtsp://localhost:8554/stream`
- [ ] Unit tests for core components

## 🚀 Quick Start

```bash
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
