# Video Streamer

Simple Go video streaming server supporting H264 over RTSP and UDP protocols.

## Quick Start

```bash
# RTSP streaming
go run server.go

# UDP streaming  
go run udp_h264_streamer.go camera_stream.h264
```

## View Streams

**RTSP:**
```bash
ffplay rtsp://localhost:8554/stream
```

**UDP:**
```bash
ffplay -f h264 -framerate 30 udp://localhost:9999
```

## Build
```bash
make build
make run
```
