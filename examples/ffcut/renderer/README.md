# FFcut renderer example

This example demonstrates the executable FFcut Project v2 path:

```text
local clips + narration -> ffcut.Project -> ffcut.Marshal -> renderer.Render -> MP4
```

Prepare two local video files and one narration file. Each video must cover at
least half of `-duration`; the narration must cover the complete duration.
FFmpeg must be available on `PATH`.

```bash
go run ./examples/ffcut/renderer \
  -clip-a /absolute/path/to/first.mp4 \
  -clip-b /absolute/path/to/second.mp4 \
  -voice /absolute/path/to/narration.wav \
  -out /absolute/path/to/final.mp4 \
  -duration 4s \
  -width 720 \
  -height 1280 \
  -debug
```

The output uses a centered cover crop, a hard cut at the midpoint, disabled
source audio, and the narration as its only audio track. The program prints the
validated Project JSON first; `-debug` additionally prints the quoted FFmpeg
command before rendering.
