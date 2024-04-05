package stream

import (
	"fmt"
)

type Stream string

const (
	Audio Stream = "a"
	Video Stream = "v"

	MayAudio Stream = "a?"
	MayVideo Stream = "v?"
)

type PosFrom string

const (
	PosFromInput  PosFrom = "input"
	PosFromFilter PosFrom = "filter"
	PosFromOutput PosFrom = "output"
)

type Streamer interface {
	Name(PosFrom) string
}

// 常量型stream
type StreamImpl string

func (s StreamImpl) Name(pf PosFrom) string {
	return string(s)
}

func Select(idx int, s Stream) Streamer {
	return StreamImpl(fmt.Sprintf("%d:%s", idx, s))
}

// 从input中选择视频流，仅用于filter中
func V(i int) Streamer {
	return StreamImpl(fmt.Sprintf("[%d:%s]", i, Video))
}

// 从input中选择音频流，仅用于filter中
func A(i int) Streamer {
	return StreamImpl(fmt.Sprintf("[%d:%s]", i, Audio))
}
