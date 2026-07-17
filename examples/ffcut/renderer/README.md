# FFcut renderer example

The renderer consumes an existing FFcut Project v2 protocol. Put a validated
project at `project.ffcut.json`, then run:

```bash
go run ./examples/ffcut/renderer
```

The example strictly decodes the project with `ffcut.Unmarshal`, prints the
FFmpeg command, and renders `final.mp4`.

All `LocalSource.path` values in the project must be absolute local paths and
must remain consistent with their fingerprints. The current renderer supports:

- a color canvas background;
- one or more cut-connected video clips;
- `cover`, `contain`, and `stretch` fit modes;
- disabled original clip audio;
- exactly one full-timeline voice track;
- no visual layers, fades, loops, or frozen frames.

Template construction and media selection do not belong in this example. The
renderer only consumes the protocol it receives.
