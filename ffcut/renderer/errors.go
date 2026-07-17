// Package renderer renders validated ffcut projects with FFmpeg.
package renderer

import "errors"

var (
	// ErrInvalidOutput reports an empty or non-absolute output path.
	ErrInvalidOutput = errors.New("ffcut renderer: invalid output")
	// ErrUnsupportedProject reports a valid Project feature the renderer does not implement.
	ErrUnsupportedProject = errors.New("ffcut renderer: unsupported project")
	// ErrRenderFailed reports an FFmpeg process failure.
	ErrRenderFailed = errors.New("ffcut renderer: render failed")
)
