import cv2

rtsp_url = "rtsp://127.0.0.1:8554/stream"

cap = cv2.VideoCapture(rtsp_url)
if not cap.isOpened():
    print("Error: Cannot open RTSP stream")
    exit()

fourcc = cv2.VideoWriter_fourcc(*"mp4v")
out = None
recording = False
paused = False

print("Press SPACE to pause/unpause")
print("Press 'r' to start/stop recording")
print("Press 'q' or ESC to quit")

while True:
    if not paused:
        ret, frame = cap.read()
        if not ret:
            print("Stream ended or error reading frame")
            break

        cv2.imshow("RTSP Stream", frame)

        if recording and out is not None:
            out.write(frame)

    key = cv2.waitKey(30) & 0xFF
    if key == 27 or key == ord('q'):  # ESC or q
        print("Exiting...")
        break
    elif key == ord(' '):  # SPACE to pause/unpause
        paused = not paused
        print("Paused" if paused else "Resumed")
    elif key == ord('r'):  # Toggle recording
        if not recording:
            # Start recording, get frame size and fps from capture
            width = int(cap.get(cv2.CAP_PROP_FRAME_WIDTH))
            height = int(cap.get(cv2.CAP_PROP_FRAME_HEIGHT))
            fps = cap.get(cv2.CAP_PROP_FPS)
            if fps == 0 or fps != fps:  # sometimes fps is zero or NaN
                fps = 25.0
            out = cv2.VideoWriter("output.mp4", fourcc, fps, (width, height))
            if not out.isOpened():
                print("Error: Cannot open video writer")
                out = None
                continue
            recording = True
            print("Recording started")
        else:
            recording = False
            out.release()
            out = None
            print("Recording stopped")

cap.release()
if out is not None:
    out.release()
cv2.destroyAllWindows()

