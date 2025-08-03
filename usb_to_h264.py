#!/usr/bin/env python3
"""
Real-time USB Camera Streaming System
Captures frames from USB camera and streams to named pipe for ffplay consumption
"""

import cv2
import os
import sys
import time
import argparse
import signal

class CameraStreamer:
    def __init__(self, 
                 camera_id: int = 0,
                 width: int = 640,
                 height: int = 480,
                 fps: int = 30,
                 pipe_path: str = "/tmp/camera_stream"):
        """
        Initialize the camera streamer
        
        Args:
            camera_id: USB camera device ID (usually 0)
            width: Frame width
            height: Frame height
            fps: Target frames per second
            pipe_path: Named pipe path for streaming
        """
        self.camera_id = camera_id
        self.width = width
        self.height = height
        self.fps = fps
        self.pipe_path = pipe_path
        
        self.cap = None
        self.pipe_fd = None
        self.running = False
        
        # Performance monitoring
        self.frame_count = 0
        self.start_time = None
        
    def setup_camera(self) -> bool:
        """Initialize and configure the USB camera"""
        print(f"Initializing camera {self.camera_id}...")
        
        self.cap = cv2.VideoCapture(self.camera_id)
        if not self.cap.isOpened():
            print(f"Error: Could not open camera {self.camera_id}")
            return False
            
        # Configure camera properties
        self.cap.set(cv2.CAP_PROP_FRAME_WIDTH, self.width)
        self.cap.set(cv2.CAP_PROP_FRAME_HEIGHT, self.height)
        self.cap.set(cv2.CAP_PROP_FPS, self.fps)
        
        # Reduce buffer size to minimize latency
        self.cap.set(cv2.CAP_PROP_BUFFERSIZE, 1)
        
        # Verify settings
        actual_width = int(self.cap.get(cv2.CAP_PROP_FRAME_WIDTH))
        actual_height = int(self.cap.get(cv2.CAP_PROP_FRAME_HEIGHT))
        actual_fps = self.cap.get(cv2.CAP_PROP_FPS)
        
        print(f"Camera configured: {actual_width}x{actual_height} @ {actual_fps} FPS")
        return True
        
    def setup_pipe(self) -> bool:
        """Create and open named pipe for streaming"""
        try:
            # Remove existing pipe if it exists
            if os.path.exists(self.pipe_path):
                os.unlink(self.pipe_path)
                
            # Create named pipe
            os.mkfifo(self.pipe_path)
            print(f"Created named pipe: {self.pipe_path}")
            
            # Open pipe for writing (this will block until reader connects)
            print("Waiting for ffplay to connect...")
            self.pipe_fd = os.open(self.pipe_path, os.O_WRONLY)
            print("ffplay connected!")
            
            return True
            
        except Exception as e:
            print(f"Error setting up pipe: {e}")
            return False
            
    def start_streaming(self) -> bool:
        """Start the streaming process"""
        if not self.setup_camera():
            return False
            
        if not self.setup_pipe():
            return False
            
        print("Starting streaming...")
        self.running = True
        self.start_time = time.time()
        
        frame_interval = 1.0 / self.fps
        last_frame_time = 0
        last_status_time = time.time()
        
        try:
            while self.running:
                current_time = time.time()
                
                # Maintain target FPS
                if current_time - last_frame_time < frame_interval:
                    time.sleep(0.001)  # Small sleep to prevent busy waiting
                    continue
                
                ret, frame = self.cap.read()
                if not ret:
                    print("Warning: Failed to capture frame")
                    continue
                
                # Encode frame as JPEG for streaming
                ret, buffer = cv2.imencode('.jpg', frame, 
                                         [cv2.IMWRITE_JPEG_QUALITY, 80])
                
                if ret and self.pipe_fd is not None:
                    try:
                        os.write(self.pipe_fd, buffer.tobytes())
                        self.frame_count += 1
                        last_frame_time = current_time
                        
                    except BrokenPipeError:
                        print("ffplay disconnected")
                        break
                    except Exception as e:
                        print(f"Error writing to pipe: {e}")
                        break
                
                # Print status every 5 seconds
                if current_time - last_status_time >= 5.0:
                    elapsed = current_time - self.start_time
                    avg_fps = self.frame_count / elapsed if elapsed > 0 else 0
                    print(f"Status: {avg_fps:.1f} FPS, Total frames: {self.frame_count}")
                    last_status_time = current_time
                        
        except KeyboardInterrupt:
            pass
            
        return True
        
    def stop_streaming(self):
        """Stop streaming and cleanup resources"""
        print("\nStopping stream...")
        self.running = False
        
        # Cleanup resources
        if self.cap:
            self.cap.release()
            
        if self.pipe_fd:
            os.close(self.pipe_fd)
            
        if os.path.exists(self.pipe_path):
            os.unlink(self.pipe_path)
            
        # Print statistics
        if self.start_time:
            duration = time.time() - self.start_time
            avg_fps = self.frame_count / duration if duration > 0 else 0
            print(f"Streaming statistics:")
            print(f"  Duration: {duration:.1f} seconds")
            print(f"  Frames streamed: {self.frame_count}")
            print(f"  Average FPS: {avg_fps:.1f}")


def signal_handler(signum, frame):
    """Handle Ctrl+C gracefully"""
    print("\nReceived interrupt signal...")
    global streamer
    if streamer:
        streamer.stop_streaming()
    sys.exit(0)


def main():
    parser = argparse.ArgumentParser(description='USB Camera Streaming System')
    parser.add_argument('--camera', '-c', type=int, default=0,
                       help='Camera device ID (default: 0)')
    parser.add_argument('--width', '-w', type=int, default=640,
                       help='Frame width (default: 640)')
    parser.add_argument('--height', '-H', type=int, default=480,
                       help='Frame height (default: 480)')
    parser.add_argument('--fps', '-f', type=int, default=30,
                       help='Target FPS (default: 30)')
    parser.add_argument('--pipe', '-p', type=str, default='/tmp/camera_stream',
                       help='Named pipe path (default: /tmp/camera_stream)')
    
    args = parser.parse_args()
    
    global streamer
    streamer = CameraStreamer(
        camera_id=args.camera,
        width=args.width,
        height=args.height,
        fps=args.fps,
        pipe_path=args.pipe
    )
    
    # Set up signal handler for graceful shutdown
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)
    
    print("=== USB Camera Streaming System ===")
    print(f"Camera: {args.camera}")
    print(f"Resolution: {args.width}x{args.height}")
    print(f"Target FPS: {args.fps}")
    print(f"Pipe: {args.pipe}")
    print("Press Ctrl+C to stop\n")
    
    if streamer.start_streaming():
        print(f"\nTo view stream, run in another terminal:")
        print(f"ffplay -f mjpeg -i {args.pipe}")
        print(f"# or")
        print(f"ffplay -f mjpeg -fflags nobuffer -flags low_delay -i {args.pipe}")
    else:
        print("Failed to start streaming")
        sys.exit(1)


if __name__ == "__main__":
    main()