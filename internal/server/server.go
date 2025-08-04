package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"matek-video-streamer/internal/streamer"
	"matek-video-streamer/internal/utils"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/description"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
)

func StartServer(videoFilePath, rtspAddress, udpRTPAddress, udpRTCPAddress string) error {
	h := NewHandler()

	// prevent clients from connecting to the server until the stream is properly set up
	h.Lock()

	// load certificates - they can be generated with
	// openssl genrsa -out server.key 2048
	// openssl req -new -x509 -sha256 -key server.key -out server.crt -days 3650
	cert, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		panic(err)
	}

	// create the server
	rtspServer := &gortsplib.Server{
		Handler:           h,
		TLSConfig:         &tls.Config{Certificates: []tls.Certificate{cert}},
		RTSPAddress:       rtspAddress,
		UDPRTPAddress:     udpRTPAddress,
		UDPRTCPAddress:    udpRTCPAddress,
		MulticastIPRange:  "224.1.0.0/16",
		MulticastRTPPort:  8002,
		MulticastRTCPPort: 8003,
	}

	h.SetServer(rtspServer)

	// start the server
	err = rtspServer.Start()
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	defer rtspServer.Close()

	// Extract H.264 parameters (SPS/PPS) from the video file
	h264Params, err := utils.ExtractH264ParametersFromHex(videoFilePath)
	if err != nil {
		log.Printf("Warning: Failed to extract H.264 parameters using hex method: %v", err)
		// Try alternative method
		h264Params, err = utils.ExtractH264Parameters(videoFilePath)
		if err != nil {
			log.Printf("ERROR: Failed to extract H.264 parameters: %v", err)
			// Fallback to basic configuration without SPS/PPS
			h264Params = nil
		}
	}

	var h264Format format.Format
	if h264Params != nil {
		log.Printf("Successfully extracted SPS (%d bytes) and PPS (%d bytes)", len(h264Params.SPS), len(h264Params.PPS))
		// Create H.264 format with SPS and PPS
		h264Format = &format.H264{
			PayloadTyp:        96,
			PacketizationMode: 1,
			SPS:               h264Params.SPS,
			PPS:               h264Params.PPS,
		}
	} else {
		log.Printf("Using basic H.264 configuration without SPS/PPS")
		// Fallback configuration
		h264Format = &format.H264{
			PayloadTyp:        96,
			PacketizationMode: 1,
		}
	}

	// create a RTSP description that contains a H264 format with SPS/PPS
	desc := &description.Session{
		Medias: []*description.Media{{
			Type:    description.MediaTypeVideo,
			Formats: []format.Format{h264Format},
		}},
	}

	// create a server stream
	stream := &gortsplib.ServerStream{
		Server: rtspServer,
		Desc:   desc,
	}
	err = stream.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize stream: %w", err)
	}
	defer stream.Close()

	h.SetStream(stream)

	// create file streamer
	r := streamer.NewFileStreamer(stream, videoFilePath)
	err = r.Initialize()
	if err != nil {
		return fmt.Errorf("failed to initialize file streamer: %w", err)
	}
	defer r.Close()

	// allow clients to connect
	h.Unlock()

	// wait until a fatal error
	log.Printf("server is ready on %s", rtspServer.RTSPAddress)
	return fmt.Errorf("server stopped: %w", rtspServer.Wait())
}
