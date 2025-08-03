package streamer

import (
	"fmt"
	"os"
	"strings"

	"github.com/bluenviron/gortsplib/v4"
)

type FileStreamer interface {
	Initialize() error
	Close() error
	Stream() *gortsplib.ServerStream
}

func NewFileStreamer(stream *gortsplib.ServerStream, filePath string) FileStreamer {
	// Check if input file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		panic(fmt.Sprintf("Input file does not exist: %s\n", filePath))
	}

	// check if the file is in MPEG-TS format
	if strings.HasSuffix(filePath, ".ts") {

		// open a file in MPEG-TS format
		f, err := os.Open(filePath)
		if err != nil {
			panic(err)
		}
		//TODO: a validation step to ensure the file is indeed MPEG-TS
		// reset the file pointer to the beginning
		f.Seek(0, 0)

		// create a new file streamer
		return &mjpegtsFileStreamer{
			stream: stream,
			f:      f,
		}
	}

	if strings.HasSuffix(filePath, ".mp4") {
		// open a file in MPEG-TS format
		f, err := os.Open(filePath)
		if err != nil {
			panic(err)
		}
		//TODO: a validation step to ensure the file is indeed MP4
		// reset the file pointer to the beginning
		f.Seek(0, 0)

		// create a new file streamer
		return &mp4FileStreamer{
			stream: stream,
			f:      f,
		}
	}

	// create a new file streamer
	// open a named pipe for MJPEG
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	return &mjpegStreamer{
		stream: stream,
		f:      file,
	}
}
