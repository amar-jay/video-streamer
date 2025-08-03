package streamer

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/bluenviron/gortsplib/v4"
	"github.com/bluenviron/gortsplib/v4/pkg/format"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts"
	"github.com/pion/rtp"
)

// mjpegtsFileStreamer reads an MPEG-TS file and streams its content as H264 over RTP.
type mjpegtsFileStreamer struct {
	stream *gortsplib.ServerStream
	f      *os.File
}

func (r *mjpegtsFileStreamer) Initialize() error {
	// check if the file is opened
	if r.f == nil {
		return os.ErrInvalid
	}

	// in a separate routine, route frames from file to ServerStream
	go r.run()

	return nil
}

func (r *mjpegtsFileStreamer) Stream() *gortsplib.ServerStream {
	return r.stream
}

func (r *mjpegtsFileStreamer) Close() error {
	// close the file
	if r.f != nil {
		r.f.Close()
		r.f = nil
	}
	return nil
}

func randUint32() (uint32, error) {
	var b [4]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return 0, err
	}
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3]), nil
}

func findTrack(r *mpegts.Reader) (*mpegts.Track, error) {
	for _, track := range r.Tracks() {
		if _, ok := track.Codec.(*mpegts.CodecH264); ok {
			return track, nil
		}
	}
	return nil, fmt.Errorf("H264 track not found")
}

func (r *mjpegtsFileStreamer) initialize() error {
	// open a file in MPEG-TS format
	var err error
	r.f, err = os.Open("myvideo.ts")
	if err != nil {
		return err
	}

	// in a separate routine, route frames from file to ServerStream
	go r.run()

	return nil
}

func (r *mjpegtsFileStreamer) close() {
	r.f.Close()
}

func (r *mjpegtsFileStreamer) run() {
	// setup H264 -> RTP encoder
	rtpEnc, err := r.stream.Desc.Medias[0].Formats[0].(*format.H264).CreateEncoder()
	if err != nil {
		panic(err)
	}

	randomStart, err := randUint32()
	if err != nil {
		panic(err)
	}

	// Check if the H.264 format has SPS/PPS and send them first
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

	for {
		// setup MPEG-TS parser
		mr := &mpegts.Reader{R: r.f}
		err = mr.Initialize()
		if err != nil {
			panic(err)
		}

		// find the H264 track inside the file
		var track *mpegts.Track
		track, err = findTrack(mr)
		if err != nil {
			panic(err)
		}

		timeDecoder := mpegts.TimeDecoder{}
		timeDecoder.Initialize()

		var firstDTS *int64
		var firstTime time.Time
		var lastRTPTime uint32
		var foundIDR bool = false

		// setup a callback that is called when a H264 access unit is read from the file
		mr.OnDataH264(track, func(pts, dts int64, au [][]byte) error {
			dts = timeDecoder.Decode(dts)
			pts = timeDecoder.Decode(pts)

			// Check if this access unit contains an IDR frame
			isIDR := false
			for _, nalUnit := range au {
				if len(nalUnit) > 0 {
					nalType := nalUnit[0] & 0x1F
					if nalType == 5 { // IDR frame
						isIDR = true
						break
					}
				}
			}

			// Skip frames until we find the first IDR frame
			if !foundIDR {
				if !isIDR {
					log.Printf("Skipping non-IDR frame (NAL type: %d), waiting for IDR", au[0][0]&0x1F)
					return nil // Skip this frame
				}
				foundIDR = true
				log.Printf("Found IDR frame, starting stream transmission")
			}

			// sleep between access units
			if firstDTS != nil {
				timeDrift := time.Duration(dts-*firstDTS)*time.Second/90000 - time.Since(firstTime)
				if timeDrift > 0 {
					time.Sleep(timeDrift)
				}
			} else {
				firstTime = time.Now()
				firstDTS = &dts
			}

			// Log frame type for debugging
			// frameType := "Unknown"
			// if len(au) > 0 && len(au[0]) > 0 {
			// 	nalType := au[0][0] & 0x1F
			// 	switch nalType {
			// 	case 1:
			// 		frameType = "P-frame"
			// 	case 5:
			// 		frameType = "IDR"
			// 	case 7:
			// 		frameType = "SPS"
			// 	case 8:
			// 		frameType = "PPS"
			// 	default:
			// 		frameType = fmt.Sprintf("NAL-%d", nalType)
			// 	}
			// }
			// log.Printf("Sending %s with pts=%d dts=%d", frameType, pts, dts)

			// wrap the access unit into RTP packets
			var packets []*rtp.Packet
			packets, err = rtpEnc.Encode(au)
			if err != nil {
				return err
			}

			// set packet timestamp
			// we don't have to perform any conversion
			// since H264 clock rate is the same in both MPEG-TS and RTSP
			lastRTPTime = uint32(int64(randomStart) + pts)
			for _, packet := range packets {
				packet.Timestamp = lastRTPTime
			}

			// write RTP packets to the server
			for _, packet := range packets {
				err = r.stream.WritePacketRTP(r.stream.Desc.Medias[0], packet)
				if err != nil {
					return err
				}
			}

			return nil
		})

		// read the file
		for {
			err = mr.Read()
			if err != nil {
				// file has ended
				if errors.Is(err, io.EOF) {
					log.Printf("file has ended, rewinding")

					// rewind to start position
					_, err = r.f.Seek(0, io.SeekStart)
					if err != nil {
						panic(err)
					}

					// Reset foundIDR flag to ensure we start with IDR again
					foundIDR = false

					// keep current timestamp
					randomStart = lastRTPTime + 1

					break
				}
				panic(err)
			}
		}
	}
}
