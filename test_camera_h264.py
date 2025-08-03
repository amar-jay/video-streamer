#!/usr/bin/env python3
"""
Test script for the camera to H.264 converter.
This script demonstrates how to use the CameraToH264Converter class.
"""

import sys
import os
from pathlib import Path

# Add the internal utils directory to the path
current_dir = os.path.dirname(os.path.abspath(__file__))
utils_path = os.path.join(current_dir, 'internal', 'utils')
sys.path.insert(0, utils_path)

try:
    from usb_to_h264 import CameraToH264Converter
except ImportError as e:
    print(f"Failed to import CameraToH264Converter: {e}")
    print("Please ensure the usb_to_h264.py file exists in internal/utils/")
    sys.exit(1)

def test_short_recording():
    """Test recording a short video clip."""
    print("Testing short recording (10 seconds)...")
    
    converter = CameraToH264Converter(
        camera_index=0,
        output_file="test_output.h264",
        fps=30,
        width=640,
        height=480
    )
    
    # Record for 10 seconds with preview
    converter.start_recording(duration=10, show_preview=True)
    
    # Check if file was created
    if Path("test_output.h264").exists():
        print("✓ H.264 file created successfully!")
        file_size = Path("test_output.h264").stat().st_size
        print(f"File size: {file_size} bytes")
    else:
        print("✗ Failed to create H.264 file")

def test_custom_settings():
    """Test with custom settings."""
    print("\nTesting with custom settings...")
    
    converter = CameraToH264Converter(
        camera_index=0,
        output_file="custom_output.h264",
        fps=24,
        width=1280,
        height=720
    )
    
    # Record for 5 seconds
    converter.start_recording(duration=5, show_preview=False)

if __name__ == "__main__":
    print("Camera to H.264 Converter Test")
    print("=" * 40)
    
    try:
        test_short_recording()
        test_custom_settings()
        print("\n✓ All tests completed!")
    except (RuntimeError, OSError, ImportError) as e:
        print(f"✗ Test failed: {e}")
