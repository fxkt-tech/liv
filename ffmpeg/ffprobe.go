package ffmpeg

import (
	"context"
	"encoding/json"
	"os/exec"
)

type FFprobeOption func(*FFprobe)

func FPCmdLoc(loc string) FFprobeOption {
	return func(f *FFprobe) {
		f.cmd = loc
	}
}

type FFprobe struct {
	cmd          string
	v            LogLevel // loglevel
	print_format string
	show_format  bool
	show_streams bool
	input        string

	probe    *Probe
	Sentence string
}

func NewProbe(opts ...FFprobeOption) *FFprobe {
	f := &FFprobe{
		cmd:          "ffprobe",
		v:            LogLevelQuiet,
		print_format: "json",
		show_format:  true,
		show_streams: true,
	}
	for _, o := range opts {
		o(f)
	}
	return f
}

func (ff *FFprobe) CmdLoc(loc string) {
	ff.cmd = loc
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

func (ff *FFprobe) SetInput(input string) {
	ff.input = input
}

func (ff *FFprobe) Run(ctx context.Context) (err error) {
	cc := exec.CommandContext(ctx, ff.cmd, ff.Params()...)
	ff.Sentence = cc.String()
	retbytes, err := cc.CombinedOutput()
	if err != nil {
		return err
	}
	probe := &Probe{}
	err = json.Unmarshal(retbytes, probe)
	if err != nil {
		return
	}
	ff.probe = probe
	return
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
