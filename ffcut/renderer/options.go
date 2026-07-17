package renderer

import (
	"context"
	"io"
	"os/exec"
)

const (
	defaultFFmpegBin    = "ffmpeg"
	defaultVideoCRF     = 20
	defaultAudioRate    = 48000
	defaultAudioBitrate = "192k"
)

type commandRunner interface {
	Run(ctx context.Context, bin string, args []string) ([]byte, error)
}

type execRunner struct{}

func (execRunner) Run(ctx context.Context, bin string, args []string) ([]byte, error) {
	return exec.CommandContext(ctx, bin, args...).CombinedOutput()
}

type config struct {
	ffmpegBin    string
	videoCRF     int
	audioRate    int
	audioBitrate string
	debugWriter  io.Writer
	runner       commandRunner
}

func defaultConfig() config {
	return config{
		ffmpegBin:    defaultFFmpegBin,
		videoCRF:     defaultVideoCRF,
		audioRate:    defaultAudioRate,
		audioBitrate: defaultAudioBitrate,
		runner:       execRunner{},
	}
}

// Option configures a Render call.
type Option func(*config)

// WithFFmpegBin selects the FFmpeg executable.
func WithFFmpegBin(bin string) Option {
	return func(config *config) {
		if bin != "" {
			config.ffmpegBin = bin
		}
	}
}

// WithVideoCRF selects the libx264 constant-rate factor.
func WithVideoCRF(crf int) Option {
	return func(config *config) {
		if crf >= 0 && crf <= 51 {
			config.videoCRF = crf
		}
	}
}

// WithDebug writes the executable and quoted arguments before running it.
func WithDebug(writer io.Writer) Option {
	return func(config *config) {
		config.debugWriter = writer
	}
}

func withRunner(runner commandRunner) Option {
	return func(config *config) {
		if runner != nil {
			config.runner = runner
		}
	}
}
