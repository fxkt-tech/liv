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

type Streamer interface {
	Name() string
}

type StreamImpl string

func (s StreamImpl) Name() string {
	return string(s)
}

func Select(idx int, s Stream) Streamer {
	return StreamImpl(fmt.Sprintf("%d:%s", idx, s))
}

// 从input中选择视频流，仅用于filter中
func V(i int) Streamer {
	return StreamImpl(fmt.Sprintf("[%d:v]", i))
}

// 从input中选择音频流，仅用于filter中
func A(i int) Streamer {
	return StreamImpl(fmt.Sprintf("[%d:a]", i))
}
