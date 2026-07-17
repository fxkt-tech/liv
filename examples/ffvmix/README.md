# FFVMix example

The example directly demonstrates the FFVMix object model:

```text
Template -> Slot / Join / BGM / Layer -> CompiledTemplate -> Generator -> FFcut Project
```

Prepare this local asset tree:

```text
/absolute/path/to/assets/
├── audio/
│   ├── bgm-a.wav
│   └── bgm-b.wav
├── image/
│   └── logo.png
└── video/
    ├── opening-a.mp4
    ├── opening-b.mp4
    ├── body-a.mp4
    └── body-b.mp4
```

Then run:

```bash
FFVMIX_ASSET_DIR=/absolute/path/to/assets go run ./examples/ffvmix
```

The program prints the persisted FFVMix template followed by the first three
validated FFcut project v2 results. Three is only the caller's take count; it
is not stored as a generator constraint.

The opening slot has a fixed five-second duration: longer sources are
center-trimmed and shorter sources are looped. The body slot omits its target
duration and keeps each selected video's natural duration.

The template also contains two transition candidates, a global BGM pool,
watermark and structured subtitle layers, built-in history constraints, and an
inline custom constraint plugin.
