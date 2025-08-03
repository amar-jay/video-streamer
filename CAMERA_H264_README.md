# USB Camera to H.264 Video Converter

This script captures video from a USB camera, processes it (converts to grayscale), and writes it to an H.264 file that can be read by other applications for display.

## Features

- Captures video from USB camera using OpenCV
- Converts frames to grayscale for processing
- Outputs H.264 encoded video files
- Configurable resolution, FPS, and duration
- Optional live preview window
- Command-line interface with various options

## Dependencies

Install the required dependencies:

```bash
pip install -r requirements.txt
```

Required packages:
- opencv-python >= 4.5.0
- numpy >= 1.19.0

## Usage

### Basic Usage

```bash
# Record indefinitely (press 'q' to stop if preview is enabled)
python3 internal/utils/usb_to_h264.py

# Record for 30 seconds
python3 internal/utils/usb_to_h264.py --duration 30

# Record with live preview
python3 internal/utils/usb_to_h264.py --preview --duration 10
```

### Advanced Usage

```bash
# Custom resolution and FPS
python3 internal/utils/usb_to_h264.py \
    --width 1280 \
    --height 720 \
    --fps 60 \
    --output my_video.h264

# Use different camera (if multiple cameras available)
python3 internal/utils/usb_to_h264.py --camera 1

# Record HD video for 2 minutes with preview
python3 internal/utils/usb_to_h264.py \
    --width 1920 \
    --height 1080 \
    --fps 30 \
    --duration 120 \
    --preview \
    --output hd_recording.h264
```

### Command Line Options

- `--camera, -c`: Camera index (default: 0)
- `--output, -o`: Output H.264 file path (default: output.h264)
- `--fps, -f`: Frames per second (default: 30)
- `--width, -w`: Video width (default: 640)
- `--height, -h`: Video height (default: 480)
- `--duration, -d`: Recording duration in seconds (default: unlimited)
- `--preview, -p`: Show live preview window

## Programmatic Usage

You can also use the `CameraToH264Converter` class directly in your Python code:

```python
from internal.utils.usb_to_h264 import CameraToH264Converter

# Create converter instance
converter = CameraToH264Converter(
    camera_index=0,
    output_file="my_video.h264",
    fps=30,
    width=640,
    height=480
)

# Start recording for 30 seconds with preview
converter.start_recording(duration=30, show_preview=True)
```

## Output Format

The script outputs H.264 encoded video files that can be:
- Played by media players (VLC, ffplay, etc.)
- Streamed via RTSP servers
- Processed by other video applications
- Converted to other formats using ffmpeg

### Playing the Output

```bash
# Play with VLC
vlc output.h264

# Play with ffplay
ffplay output.h264

# Convert to MP4 for better compatibility
ffmpeg -i output.h264 -c copy output.mp4
```

## Testing

Run the test script to verify everything works:

```bash
python3 test_camera_h264.py
```

This will:
1. Record a 10-second test video with preview
2. Record a 5-second video with custom settings
3. Verify the output files were created

## Troubleshooting

### Camera Not Found
- Check if your camera is connected and working
- Try different camera indices (0, 1, 2, etc.)
- Ensure no other application is using the camera

### Permission Issues
- On Linux, you might need to add your user to the `video` group:
  ```bash
  sudo usermod -a -G video $USER
  ```
- Restart your session after adding to the group

### Video Codec Issues
- Ensure OpenCV was compiled with H.264 support
- If H.264 encoding fails, try installing additional codecs:
  ```bash
  sudo apt-get install libx264-dev
  ```

### Performance Issues
- Lower the resolution or FPS for better performance
- Close the preview window if not needed (improves performance)
- Ensure sufficient disk space for video files

## Integration with RTSP Server

The generated H.264 files can be streamed using the Go RTSP server in this project:

```bash
# Start the RTSP server (from project root)
go run server.go

# Stream the recorded video
# (Configure the server to use your H.264 file)
```
