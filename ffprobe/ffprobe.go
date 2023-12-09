package ffprobe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/fxkt-tech/liv/ffmpeg"
)

type Option func(*FFprobe)

func ProbeBinary(bin string) Option {
	return func(f *FFprobe) {
		f.bin = bin
	}
}

func ProbeDebug(debug bool) Option {
	return func(f *FFprobe) {
		f.debug = debug
	}
}

type FFprobe struct {
	dry bool // dry run

	debug bool

	bin          string
	v            ffmpeg.LogLevel // loglevel
	print_format string
	show_format  bool
	show_streams bool
	input        string

	probe    *Probe
	Sentence string
}

func New(opts ...Option) *FFprobe {
	f := &FFprobe{
		bin:          "ffprobe",
		v:            ffmpeg.LogLevelQuiet,
		print_format: "json",
		show_format:  true,
		show_streams: true,
	}
	for _, o := range opts {
		o(f)
	}
	return f
}

func (ff *FFprobe) Params() (params []string) {
	params = append(params, "-v", ff.v.String())
	params = append(params, "-print_format", ff.print_format)
	if ff.show_format {
		params = append(params, "-show_format")
	}
	if ff.show_streams {
		params = append(params, "-show_streams")
	}
	params = append(params, ff.input)
	return
}

func (ff *FFprobe) Input(input string) *FFprobe {
	ff.input = input
	return ff
}

func (ff *FFprobe) DryRun() {
	var ps []string
	ps = append(ps, ff.bin)
	ps = append(ps, ff.Params()...)
	fmt.Println(strings.Join(ps, " "))
}

func (ff *FFprobe) Run(ctx context.Context) error {
	if ff.debug {
		ff.DryRun()
	} else {
		if ff.dry {
			ff.DryRun()
			return nil
		}
	}

	cc := exec.CommandContext(ctx, ff.bin, ff.Params()...)
	ff.Sentence = cc.String()
	retbytes, err := cc.CombinedOutput()
	if err != nil {
		return err
	}
	probe := &Probe{}
	err = json.Unmarshal(retbytes, probe)
	if err != nil {
		return err
	}
	if string(retbytes) == "{}" {
		return errors.New("file is not a media stream")
	}
	ff.probe = probe
	return nil
}

func (ff *FFprobe) RunRetRaw(ctx context.Context) ([]byte, error) {
	cc := exec.CommandContext(ctx, ff.bin, ff.Params()...)
	ff.Sentence = cc.String()
	retbytes, err := cc.CombinedOutput()
	if err != nil {
		return nil, err
	}
	probe := &Probe{}
	err = json.Unmarshal(retbytes, probe)
	if err != nil {
		return nil, err
	}
	if string(retbytes) == "{}" {
		return nil, errors.New("file is not a media stream")
	}
	ff.probe = probe
	return retbytes, nil
}

func (ff *FFprobe) GetFirstVideoStream() *ProbeStream {
	if ff.probe == nil {
		return nil
	}
	for _, stream := range ff.probe.Streams {
		if stream.CodecType == "video" {
			return stream
		}
	}
	return nil
}

func (ff *FFprobe) GetFirstAudioStream() *ProbeStream {
	if ff.probe == nil {
		return nil
	}
	for _, stream := range ff.probe.Streams {
		if stream.CodecType == "audio" {
			return stream
		}
	}
	return nil
}

func (ff *FFprobe) GetFormat() *ProbeFormat {
	if ff.probe == nil {
		return nil
	}
	return ff.probe.Format
}
