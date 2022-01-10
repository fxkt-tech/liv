package ffmpeg

import (
	"context"
	"encoding/json"
	"os/exec"
)

type FFprobe struct {
	cmd          string
	v            string // loglevel
	print_format string
	show_format  bool
	show_streams bool
	input        string

	probe    *Probe
	Sentence string
}

func DefaultProbe() *FFprobe {
	return &FFprobe{
		cmd:          "ffprobe",
		v:            "quiet",
		print_format: "json",
		show_format:  true,
		show_streams: true,
	}
}

func (ff *FFprobe) CmdLoc(loc string) {
	ff.cmd = loc
}

func (ff *FFprobe) Params() (params []string) {
	params = append(params, "-v", ff.v)
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
