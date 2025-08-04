package server

import (
	"fmt"
	"log"
	"sync"

	"github.com/pion/rtp"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/base"
	"github.com/bluenviron/gortsplib/v4/pkg/description"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
)

// Handler represents the RTSP server handler
type Handler struct {
	server    *gortsplib.Server
	mutex     sync.RWMutex
	stream    *gortsplib.ServerStream
	publisher *gortsplib.ServerSession
}

// NewHandler creates a new server handler
func NewHandler() *Handler {
	return &Handler{}
}

// SetServer sets the RTSP server instance
func (h *Handler) SetServer(server *gortsplib.Server) {
	h.server = server
}

// SetStream sets the server stream
func (h *Handler) SetStream(stream *gortsplib.ServerStream) {
	h.stream = stream
}

// GetStream returns the current server stream
func (h *Handler) GetStream() *gortsplib.ServerStream {
	return h.stream
}

// Lock locks the handler mutex
func (h *Handler) Lock() {
	h.mutex.Lock()
}

// Unlock unlocks the handler mutex
func (h *Handler) Unlock() {
	h.mutex.Unlock()
}

// OnConnOpen is called when a connection is opened
func (h *Handler) OnConnOpen(_ *gortsplib.ServerHandlerOnConnOpenCtx) {
	log.Printf("conn opened")
}

// OnConnClose is called when a connection is closed
func (h *Handler) OnConnClose(ctx *gortsplib.ServerHandlerOnConnCloseCtx) {
	log.Printf("conn closed (%v)", ctx.Error)
}

// OnSessionOpen is called when a session is opened
func (h *Handler) OnSessionOpen(_ *gortsplib.ServerHandlerOnSessionOpenCtx) {
	log.Printf("session opened")
}

// OnSessionClose is called when a session is closed
func (h *Handler) OnSessionClose(ctx *gortsplib.ServerHandlerOnSessionCloseCtx) {
	log.Printf("session closed")

	h.mutex.Lock()
	defer h.mutex.Unlock()

	// if the session is the publisher,
	// close the stream and disconnect any reader.
	if h.stream != nil && ctx.Session == h.publisher {
		h.stream.Close()
		h.stream = nil
	}
}

// OnDescribe is called when receiving a DESCRIBE request
func (h *Handler) OnDescribe(
	_ *gortsplib.ServerHandlerOnDescribeCtx,
) (*base.Response, *gortsplib.ServerStream, error) {
	log.Printf("DESCRIBE request")

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// no one is publishing yet
	if h.stream == nil {
		return &base.Response{
			StatusCode: base.StatusNotFound,
		}, nil, nil
	}

	// send medias that are being published to the client
	return &base.Response{
		StatusCode: base.StatusOK,
	}, h.stream, nil
}

// OnAnnounce is called when receiving an ANNOUNCE request
func (h *Handler) OnAnnounce(ctx *gortsplib.ServerHandlerOnAnnounceCtx) (*base.Response, error) {
	log.Printf("ANNOUNCE request")

	h.mutex.Lock()
	defer h.mutex.Unlock()

	// disconnect existing publisher
	if h.stream != nil {
		h.stream.Close()
		h.publisher.Close()
	}

	if h.publisher != nil {
		h.stream.Close()
		h.publisher.Close()
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
	h.stream = &gortsplib.ServerStream{
		Server: h.server,
		Desc:   ctx.Description,
	}

	err := h.stream.Initialize()
	if err != nil {
		panic(err)
	}

	h.publisher = ctx.Session

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

// OnSetup is called when receiving a SETUP request
func (h *Handler) OnSetup(ctx *gortsplib.ServerHandlerOnSetupCtx) (
	*base.Response, *gortsplib.ServerStream, error,
) {
	log.Printf("SETUP request")

	// SETUP is used by both readers and publishers. In case of publishers, just return StatusOK.
	if ctx.Session.State() == gortsplib.ServerSessionStatePreRecord {
		return &base.Response{
			StatusCode: base.StatusOK,
		}, nil, nil
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// no one is publishing yet
	if h.stream == nil {
		return &base.Response{
			StatusCode: base.StatusNotFound,
		}, nil, nil
	}

	return &base.Response{
		StatusCode: base.StatusOK,
	}, h.stream, nil
}

// OnPlay is called when receiving a PLAY request
func (h *Handler) OnPlay(_ *gortsplib.ServerHandlerOnPlayCtx) (*base.Response, error) {
	log.Printf("PLAY request")

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}

// OnRecord is called when receiving a RECORD request
func (h *Handler) OnRecord(ctx *gortsplib.ServerHandlerOnRecordCtx) (*base.Response, error) {
	log.Printf("RECORD request")

	// called when receiving a RTP packet
	ctx.Session.OnPacketRTPAny(func(medi *description.Media, _ format.Format, pkt *rtp.Packet) {
		// forward the packet to other clients (no recording)
		err := h.stream.WritePacketRTP(medi, pkt)
		if err != nil {
			log.Printf("ERR: %v", err)
		}
	})

	return &base.Response{
		StatusCode: base.StatusOK,
	}, nil
}
