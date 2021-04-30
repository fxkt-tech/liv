package ffmpeg

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type FFmpeg struct {
	cmd     string
	infiles []string
	filters []*Filter
	outputs []*Output

	Sentence string
}

func NewFFmpeg() *FFmpeg {
	return &FFmpeg{
		cmd: "ffmpeg",
	}
}

func (ff *FFmpeg) ChangeCmd(cmd string) {
	ff.cmd = cmd
}

func (ff *FFmpeg) AddInputs(inputs ...string) {
	ff.infiles = append(ff.infiles, inputs...)
}

func (ff *FFmpeg) AddFilter(filters ...*Filter) {
	ff.filters = append(ff.filters, filters...)
}

func (ff *FFmpeg) OutputGraph(output ...*Output) {
	ff.outputs = output
}

func (ff *FFmpeg) combination() []string {
	var params []string
	params = append(params, "-y")
	for _, infile := range ff.infiles {
		params = append(params, "-i", infile)
	}

	var filtersStr []string
	for _, filter := range ff.filters {
		filtersStr = append(filtersStr, filter.String())
	}
	params = append(params, "-filter_complex", strings.Join(filtersStr, ";"))
	for _, op := range ff.outputs {
		params = append(params, op.params...)
	}
	return params
}

func (ff *FFmpeg) Run() error {
	params := ff.combination()
	cmd := exec.CommandContext(context.Background(), ff.cmd, params...)
	ff.Sentence = cmd.String()
	fmt.Println(ff.Sentence)
	outBytes, err := cmd.CombinedOutput()
	fmt.Println(string(outBytes))

	if err != nil || cmd.ProcessState.ExitCode() != 0 {
		sls := strings.Split(string(outBytes), "\n")
		err = errors.New(strings.ReplaceAll(strings.Join(sls[len(sls)-4:len(sls)-1], "|"), "\n", "|"))
	}
	return err
}
