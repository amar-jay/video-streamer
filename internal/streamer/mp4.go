package streamer

import (
	"fmt"
	"matek-video-streamer/internal/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/bluenviron/gortsplib/v4"
)

type mp4FileStreamer struct {
	stream *gortsplib.ServerStream
	s      mjpegtsFileStreamer
	f      *os.File
	temp   *os.File
}

func (r *mp4FileStreamer) Initialize() error {
	// check if the file is opened
	if r.f == nil {
		return os.ErrInvalid
	}
	// Convert MP4 to TS using FFmpeg save to /tmp using input file name with .ts extension
	inputPath := r.f.Name()
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".ts"
	err := utils.MP4ToTS(inputPath, outputPath)
	if err != nil {
		return fmt.Errorf("failed to convert MP4 to TS: %w", err)
	}
	// Open the converted TS file
	r.temp, err = os.Open(outputPath)
	if err != nil {
		return fmt.Errorf("failed to open converted TS file: %w", err)
	}

	s := mjpegtsFileStreamer{
		stream: r.stream,
		f:      r.temp,
	}

	// in a separate routine, route frames from file to ServerStream
	go s.run()

	return nil
}

func (r *mp4FileStreamer) Stream() *gortsplib.ServerStream {
	return r.s.Stream()
}

func (r *mp4FileStreamer) Close() error {
	// close and delete the temporary TS file
	if r.temp != nil {
		r.temp.Close()
		os.Remove(r.temp.Name())
		r.temp = nil
	}

	// close the original MP4 file
	if r.f != nil {
		r.f.Close()
		r.f = nil
	}

	return r.s.Close()
}
