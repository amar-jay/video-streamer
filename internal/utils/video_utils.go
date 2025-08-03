package utils

import (
	"encoding/hex"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/pmp4"
)

// H264Parameters holds SPS and PPS data
type H264Parameters struct {
	SPS []byte
	PPS []byte
}

func FindH264Track(presentation *pmp4.Presentation) (*pmp4.Track, error) {
	for _, track := range presentation.Tracks {
		if track.Codec.IsVideo() {
			return track, nil
		}
	}
	return nil, fmt.Errorf("H264 track not found")
}

// ExtractH264Parameters extracts SPS and PPS from a video file using FFmpeg
func ExtractH264Parameters(filePath string) (*H264Parameters, error) {
	// Use FFmpeg to extract H.264 parameters
	cmd := exec.Command("ffmpeg",
		"-i", filePath,
		"-c:v", "copy",
		"-bsf:v", "h264_mp4toannexb",
		"-f", "h264",
		"-frames:v", "1", // Only process first frame to get SPS/PPS
		"-y",     // Overwrite output
		"pipe:1", // Output to stdout
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to extract H.264 parameters: %v", err)
	}

	return parseH264Parameters(output)
}

// ExtractH264ParametersFromHex extracts SPS and PPS using ffprobe to get hex output
func ExtractH264ParametersFromHex(filePath string) (*H264Parameters, error) {
	// Use ffprobe to get codec extradata in hex format
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-show_entries", "stream=codec_name,extradata",
		"-select_streams", "v:0",
		"-of", "csv=p=0",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to probe video file: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "h264") && strings.Contains(line, ",") {
			parts := strings.Split(line, ",")
			if len(parts) >= 2 {
				hexData := strings.TrimSpace(parts[1])
				if hexData != "" {
					data, err := hex.DecodeString(hexData)
					if err != nil {
						continue
					}
					return parseH264Parameters(data)
				}
			}
		}
	}

	return nil, fmt.Errorf("no H.264 extradata found")
}

// parseH264Parameters parses raw H.264 data to extract SPS and PPS
func parseH264Parameters(data []byte) (*H264Parameters, error) {
	params := &H264Parameters{}

	// Look for NAL units starting with 0x00000001 or 0x000001
	i := 0
	for i < len(data) {
		// Find start code
		if i+4 < len(data) && data[i] == 0x00 && data[i+1] == 0x00 && data[i+2] == 0x00 && data[i+3] == 0x01 {
			i += 4
		} else if i+3 < len(data) && data[i] == 0x00 && data[i+1] == 0x00 && data[i+2] == 0x01 {
			i += 3
		} else {
			i++
			continue
		}

		if i >= len(data) {
			break
		}

		// Get NAL unit type (first 5 bits of the byte after start code)
		nalType := data[i] & 0x1F

		// Find end of this NAL unit
		end := i + 1
		for end < len(data) {
			if end+3 < len(data) && data[end] == 0x00 && data[end+1] == 0x00 && data[end+2] == 0x01 {
				break
			}
			if end+4 < len(data) && data[end] == 0x00 && data[end+1] == 0x00 && data[end+2] == 0x00 && data[end+3] == 0x01 {
				break
			}
			end++
		}

		nalData := data[i:end]

		switch nalType {
		case 7: // SPS
			params.SPS = nalData
		case 8: // PPS
			params.PPS = nalData
		}

		i = end
	}

	if len(params.SPS) == 0 {
		return nil, fmt.Errorf("SPS not found")
	}
	if len(params.PPS) == 0 {
		return nil, fmt.Errorf("PPS not found")
	}

	return params, nil
}
