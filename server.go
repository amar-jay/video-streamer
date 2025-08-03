// Package main contains an example.
package main

import (
	"fmt"
	"log"
	"matek-video-streamer/internal/streamer"
	"matek-video-streamer/internal/utils"
	"sync"

	"github.com/pion/rtp"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/description"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
)

// This example shows how to
// 1. create a RTSP server which accepts plain connections.
// 2. allow a single client to publish a stream.
// 3. allow several clients to read the stream.

type serverHandler struct {
	server    *gortsplib.Server
	mutex     sync.RWMutex
	stream    *gortsplib.ServerStream
	publisher *gortsplib.ServerSession
}

// called when a connection is opened.
func (sh *serverHandler) OnConnOpen(_ *gortsplib.ServerHandlerOnConnOpenCtx) {
	log.Printf("conn opened")
}

// called when a connection is closed.
func (sh *serverHandler) OnConnClose(ctx *gortsplib.ServerHandlerOnConnCloseCtx) {
	log.Printf("conn closed (%v)", ctx.Error)
}

// called when a session is opened.
func (sh *serverHandler) OnSessionOpen(_ *gortsplib.ServerHandlerOnSessionOpenCtx) {
	log.Printf("session opened")
}

// called when a session is closed.
func (sh *serverHandler) OnSessionClose(ctx *gortsplib.ServerHandlerOnSessionCloseCtx) {
	log.Printf("session closed")

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	// if the session is the publisher,
	// close the stream and disconnect any reader.
	if sh.stream != nil && ctx.Session == sh.publisher {
		sh.stream.Close()
		sh.stream = nil
	}
}

// called when receiving a DESCRIBE request.
func (sh *serverHandler) OnDescribe(
	_ *gortsplib.ServerHandlerOnDescribeCtx,
) (*base.Response, *gortsplib.ServerStream, error) {
	log.Printf("DESCRIBE request")

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	// no one is publishing yet
	if sh.stream == nil {
		return &base.Response{
			StatusCode: base.StatusNotFound,
		}, nil, nil
	}

	// send medias that are being published to the client
	return &base.Response{
		StatusCode: base.StatusOK,
	}, sh.stream, nil
}

// called when receiving an ANNOUNCE request.
func (sh *serverHandler) OnAnnounce(ctx *gortsplib.ServerHandlerOnAnnounceCtx) (*base.Response, error) {
	log.Printf("ANNOUNCE request")

	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	// disconnect existing publisher
	if sh.stream != nil {
		sh.stream.Close()
		sh.publisher.Close()
	}

	if sh.publisher != nil {
		sh.stream.Close()
		sh.publisher.Close()
	}

	// find the H264 media and format
	var forma *format.H264
	medi := ctx.Description.FindFormat(&forma)
	if medi == nil {
		return &base.Response{
			StatusCode: base.StatusBadRequest,
		}, fmt.Errorf("H264 media not found")
	}

	// create the stream and save the publisher
	sh.stream = &gortsplib.ServerStream{
		Server: sh.server,
		Desc:   ctx.Description,
	}

	err := sh.stream.Initialize()
	if err != nil {
		panic(err)
	}

	sh.publisher = ctx.Session

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

// called when receiving a SETUP request.
func (sh *serverHandler) OnSetup(ctx *gortsplib.ServerHandlerOnSetupCtx) (
	*base.Response, *gortsplib.ServerStream, error,
) {
	log.Printf("SETUP request")

	// SETUP is used by both readers and publishers. In case of publishers, just return StatusOK.
	if ctx.Session.State() == gortsplib.ServerSessionStatePreRecord {
		return &base.Response{
			StatusCode: base.StatusOK,
		}, nil, nil
	}

	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	// no one is publishing yet
	if sh.stream == nil {
		return &base.Response{
			StatusCode: base.StatusNotFound,
		}, nil, nil
	}

	return &base.Response{
		StatusCode: base.StatusOK,
	}, sh.stream, nil

}

// called when receiving a PLAY request.
func (sh *serverHandler) OnPlay(_ *gortsplib.ServerHandlerOnPlayCtx) (*base.Response, error) {
	log.Printf("PLAY request")

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

// called when receiving a RECORD request.
func (sh *serverHandler) OnRecord(ctx *gortsplib.ServerHandlerOnRecordCtx) (*base.Response, error) {
	log.Printf("RECORD request")

	// called when receiving a RTP packet
	ctx.Session.OnPacketRTPAny(func(medi *description.Media, _ format.Format, pkt *rtp.Packet) {
		// forward the packet to other clients (no recording)
		err := sh.stream.WritePacketRTP(medi, pkt)
		if err != nil {
			log.Printf("ERR: %v", err)
		}
	})

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

func main() {
	h := &serverHandler{}

	// prevent clients from connecting to the server until the stream is properly set up
	h.mutex.Lock()

	// create the server
	h.server = &gortsplib.Server{
		Handler:           h,
		RTSPAddress:       ":8554",
		UDPRTPAddress:     ":8000",
		UDPRTCPAddress:    ":8001",
		MulticastIPRange:  "224.1.0.0/16",
		MulticastRTPPort:  8002,
		MulticastRTCPPort: 8003,
	}

	// start the server
	err := h.server.Start()
	if err != nil {
		panic(err)
	}
	defer h.server.Close()

	// Extract H.264 parameters (SPS/PPS) from the video file
	videoFilePath := "/home/amarjay/Downloads/demo.mp4"
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
	h.stream = &gortsplib.ServerStream{
		Server: h.server,
		Desc:   desc,
	}
	err = h.stream.Initialize()
	if err != nil {
		panic(err)
	}
	defer h.stream.Close()

	// create file streamer
	r := streamer.NewFileStreamer(h.stream, "/tmp/camera_stream")
	err = r.Initialize()
	if err != nil {
		panic(err)
	}
	defer r.Close()

	// allow clients to connect
	h.mutex.Unlock()

	// wait until a fatal error
	log.Printf("server is ready on %s", h.server.RTSPAddress)
	panic(h.server.Wait())
}
