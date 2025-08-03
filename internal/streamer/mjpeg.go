package streamer

import (
	"crypto/rand"
	"io"
	"log"
	"os"
	"time"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
)

// mjpegStreamer reads MJPEG frames from a named pipe and streams them over RTP.
type mjpegStreamer struct {
	stream *gortsplib.ServerStream
	f      *os.File
}

func (r *mjpegStreamer) Initialize() error {
	// check if the file is opened
	if r.f == nil {
		return os.ErrInvalid
	}

	// in a separate routine, route frames from file to ServerStream
	go r.run()

	return nil
}

func (r *mjpegStreamer) Stream() *gortsplib.ServerStream {
	return r.stream
}

func (r *mjpegStreamer) Close() error {
	// close the file
	if r.f != nil {
		r.f.Close()
		r.f = nil
	}
	return nil
}

// randUint32MJPEG generates a random 32-bit unsigned integer for MJPEG streamer
func randUint32MJPEG() (uint32, error) {
	var b [4]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, err
	}
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]), nil
}

// findJPEGStart finds the start of a JPEG frame (0xFF 0xD8)
func findJPEGStart(data []byte) int {
	for i := 0; i < len(data)-1; i++ {
		if data[i] == 0xFF && data[i+1] == 0xD8 {
			return i
		}
	}
	return -1
}

// findJPEGEnd finds the end of a JPEG frame (0xFF 0xD9)
func findJPEGEnd(data []byte, start int) int {
	for i := start; i < len(data)-1; i++ {
		if data[i] == 0xFF && data[i+1] == 0xD9 {
			return i + 2 // Include the end marker
		}
	}
	return -1
}

func (r *mjpegStreamer) run() {
	// setup H264 -> RTP encoder (just like mjpeg-ts.go)
	rtpEnc, err := r.stream.Desc.Medias[0].Formats[0].(*format.H264).CreateEncoder()
	if err != nil {
		panic(err)
	}

	randomStart, err := randUint32MJPEG()
	if err != nil {
		panic(err)
	}

	// Check if the H.264 format has SPS/PPS and send them first (like mjpeg-ts.go)
	h264Format := r.stream.Desc.Medias[0].Formats[0].(*format.H264)
	if len(h264Format.SPS) > 0 && len(h264Format.PPS) > 0 {
		log.Printf("Sending initial SPS/PPS parameters")

		// Create access unit with SPS and PPS
		initialAU := [][]byte{h264Format.SPS, h264Format.PPS}

		// Encode SPS/PPS into RTP packets
		packets, err := rtpEnc.Encode(initialAU)
		if err != nil {
			log.Printf("Failed to encode SPS/PPS: %v", err)
		} else {
			// Send SPS/PPS packets
			for _, packet := range packets {
				packet.Timestamp = randomStart
				err = r.stream.WritePacketRTP(r.stream.Desc.Medias[0], packet)
				if err != nil {
					log.Printf("Failed to write SPS/PPS packet: %v", err)
				}
			}
		}
	}

	log.Printf("Starting MJPEG to H.264 conversion stream from named pipe")

	buffer := make([]byte, 0, 1024*1024) // 1MB buffer
	frameBuffer := make([]byte, 4096)    // 4KB read buffer
	var frameCount uint64
	startTime := time.Now()
	var lastRTPTime uint32

	for {
		// Read data from the named pipe
		n, err := r.f.Read(frameBuffer)
		if err != nil {
			if err == io.EOF {
				log.Printf("End of MJPEG stream, waiting for more data...")
				time.Sleep(100 * time.Millisecond)
				continue
			}
			log.Printf("Error reading from MJPEG stream: %v", err)
			continue
		}

		if n == 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		// Append new data to buffer
		buffer = append(buffer, frameBuffer[:n]...)

		// Process complete JPEG frames
		for {
			// Find JPEG start marker
			jpegStart := findJPEGStart(buffer)
			if jpegStart == -1 {
				// No start marker found, keep the last few bytes in case they're part of a marker
				if len(buffer) > 10 {
					buffer = buffer[len(buffer)-10:]
				}
				break
			}

			// Remove data before JPEG start
			if jpegStart > 0 {
				buffer = buffer[jpegStart:]
			}

			// Find JPEG end marker
			jpegEnd := findJPEGEnd(buffer, 2) // Start searching after the start marker
			if jpegEnd == -1 {
				// Incomplete frame, wait for more data
				break
			}

			// Extract complete JPEG frame
			jpegFrame := make([]byte, jpegEnd)
			copy(jpegFrame, buffer[:jpegEnd])

			// Remove processed frame from buffer
			buffer = buffer[jpegEnd:]

			// Validate JPEG frame
			if len(jpegFrame) < 10 {
				log.Printf("JPEG frame too small (%d bytes), skipping", len(jpegFrame))
				continue
			}

			frameCount++
			currentTime := time.Now()

			// Log frame information periodically
			if frameCount%30 == 0 {
				duration := currentTime.Sub(startTime)
				fps := float64(frameCount) / duration.Seconds()
				log.Printf("Processed %d MJPEG frames, current frame size: %d bytes, avg FPS: %.2f",
					frameCount, len(jpegFrame), fps)
			}

			// Convert JPEG frame to H.264 access unit (like mjpeg-ts.go)
			// Treat JPEG frame as a single NAL unit
			au := [][]byte{jpegFrame}

			// Calculate timestamp based on frame rate (like mjpeg-ts.go)
			// H.264 uses a 90kHz clock rate, similar to MPEG-TS
			pts := int64(frameCount * 3000) // 90000/30 = 3000 ticks per frame
			lastRTPTime = uint32(int64(randomStart) + pts)

			// wrap the access unit into RTP packets (like mjpeg-ts.go)
			packets, err := rtpEnc.Encode(au)
			if err != nil {
				log.Printf("Error encoding frame as H.264: %v", err)
				continue
			}

			// set packet timestamp (like mjpeg-ts.go)
			for _, packet := range packets {
				packet.Timestamp = lastRTPTime
			}

			// write RTP packets to the server (like mjpeg-ts.go)
			for _, packet := range packets {
				err = r.stream.WritePacketRTP(r.stream.Desc.Medias[0], packet)
				if err != nil {
					log.Printf("Error writing RTP packet: %v", err)
					continue
				}
			}
		}

		// Prevent buffer from growing too large
		if len(buffer) > 2*1024*1024 { // 2MB limit
			log.Printf("Buffer too large (%d bytes), resetting", len(buffer))
			buffer = buffer[len(buffer)/2:] // Keep last half
		}
	}
}
