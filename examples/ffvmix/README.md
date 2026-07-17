# FFVMix example

This example builds and persists an FFVMix template, compiles its local media
with FFprobe, lazily takes the first N accepted combinations, and writes each
one as a validated FFcut project v2 JSON file.

```bash
go run ./examples/ffvmix \
  -base /absolute/path/to/assets \
  -opening video/opening-a.mp4 \
  -opening video/opening-b.mp4 \
  -body video/body-a.mp4 \
  -body video/body-b.mp4 \
  -bgm audio/bgm-a.mp3 \
  -bgm audio/bgm-b.mp3 \
  -watermark image/logo.png \
  -subtitle "FFVMix demo" \
  -count 3 \
  -seed 42 \
  -out ./ffvmix-output
```

`-opening`, `-body`, and `-bgm` may be repeated to form candidate pools.
Relative asset paths are resolved against the absolute `-base` directory.
FFprobe must be available on `PATH`.

The opening slot has a fixed five-second duration: longer sources are
center-trimmed and shorter sources are looped. The body slot has no target
duration and therefore keeps the selected video's natural duration.

`-count` only stops the caller after it receives N projects. It is not a
generation constraint: the generator remains lazy and can continue until its
combination space is exhausted. The output directory contains:

```text
template.ffvmix.json
project-001.ffcut.json
project-002.ffcut.json
...
```

The example also shows color background and canvas configuration, cut/fade
transition candidates, global BGM, optional watermark and structured subtitle
layers, built-in history constraints, and a custom pure constraint plugin.
