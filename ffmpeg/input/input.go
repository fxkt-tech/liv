package input

import (
	"fmt"
	"strings"

	"github.com/fxkt-tech/liv/ffmpeg/stream"
)

// Input is common input info.
type Input struct {
	idx int

	cv        string
	r         string
	safe      string
	ss        float32  // ss is start_time.
	t         float32  // t is duration.
	itsoffset float32  // offset
	metadata  []string // kv pair.
	f         string   // format
	i         string   // i is input file.
}

func New(opts ...Option) *Input {
	i := &Input{}
	for _, o := range opts {
		o(i)
	}
	return i
}

func WithSimple(i string) *Input {
	return &Input{i: i}
}

func WithConcat(i string) *Input {
	return &Input{i: i, f: "concat", safe: "0"}
}

func WithMetadata(i string, kvs []string) *Input {
	return &Input{i: i, metadata: kvs}
}

func WithTime(ss, t float32, i string) *Input {
	return &Input{
		ss: ss,
		t:  t,
		i:  i,
	}
}

func (i *Input) V() stream.Streamer {
	return &InputStream{input: i, s: stream.Video}
}

func (i *Input) A() stream.Streamer {
	return &InputStream{input: i, s: stream.Audio}
}

func (i *Input) MayV() stream.Streamer {
	return &InputStream{input: i, s: stream.MayVideo}
}

func (i *Input) MayA() stream.Streamer {
	return &InputStream{input: i, s: stream.MayAudio}
}

func (i *Input) Params() (params []string) {
	if i.r != "" {
		params = append(params, "-r", i.r)
	}
	if i.cv != "" {
		params = append(params, "-c:v", i.cv)
	}
	if i.safe != "" {
		params = append(params, "-safe", i.safe)
	}
	if i.ss != 0 {
		params = append(params, "-ss", fmt.Sprintf("%.6f", i.ss))
	}
	if i.t != 0 {
		params = append(params, "-t", fmt.Sprintf("%.6f", i.t))
	}
	if i.itsoffset != 0 {
		params = append(params, "-itsoffset", fmt.Sprintf("%.6f", i.itsoffset))
	}
	if i.f != "" {
		params = append(params, "-f", i.f)
	}
	params = append(params, "-i", i.i)
	for j := 0; j < len(i.metadata); j += 2 {
		params = append(params, "-metadata", fmt.Sprintf("%s=%s", i.metadata[j], i.metadata[j+1]))
	}
	return
}

func (i *Input) String() string {
	return strings.Join(i.Params(), " ")
}

//

type Inputs []*Input

func (inputs Inputs) Params() (params []string) {
	for _, input := range inputs {
		params = append(params, input.Params()...)
	}
	return
}

func (inputs Inputs) String() string {
	return strings.Join(inputs.Params(), " ")
}

func (inputs Inputs) Tidy() Inputs {
	for i, input := range inputs {
		input.idx = i
	}
	return inputs
}

// stream

// input型stream
type InputStream struct {
	input *Input
	s     stream.Stream
}

func (s *InputStream) Name(pf stream.PosFrom) string {
	switch pf {
	case stream.PosFromOutput:
		return fmt.Sprintf("%d:%s", s.input.idx, s.s)
	default:
		return fmt.Sprintf("[%d:%s]", s.input.idx, s.s)
	}
}
