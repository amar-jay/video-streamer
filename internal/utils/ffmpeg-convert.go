package utils

import (
	"fmt"
	"os/exec"
)

func MP4ToTS(inputPath, outputPath string) error {
	// Build FFmpeg command with additional parameters to ensure SPS/PPS are included
	// and force the first frame to be an IDR frame
	cmd := exec.Command("ffmpeg",
		"-i", inputPath, // Input file
		"-c:v", "libx264", // Re-encode video to ensure proper frame order
		"-preset", "ultrafast", // Fast encoding
		"-tune", "zerolatency", // Low latency tuning
		"-x264-params", "keyint=30:min-keyint=30", // Force keyframes every 30 frames
		"-force_key_frames", "expr:gte(t,0)", // Force a keyframe at the start
		"-bsf:v", "h264_mp4toannexb", // Convert H.264 bitstream from MP4 to Annex B format
		"-avoid_negative_ts", "make_zero", // Avoid negative timestamps
		"-fflags", "+genpts", // Generate presentation timestamps
		"-f", "mpegts", // Output format
		"-y",       // Overwrite output file
		outputPath, // Output file
	)

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg error: %v\nOutput: %s", err, string(output))
	}

	return nil
}
