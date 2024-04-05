package filter

import (
	"fmt"

	"github.com/fxkt-tech/liv/internal/naming"
	"golang.org/x/exp/constraints"
)

// 音频流复制成多份
func ASplit(n int) *MultiFilter {
	return &MultiFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("asplit=%d", n),
		counts:  n,
	}
}

func ATempo[T constraints.Integer | constraints.Float | string](expr T) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("atempo=%v", expr),
	}
}

// 音频帧显示时间戳
func ASetPTS(expr string) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("asetpts=%s", expr),
	}
}

func AMix(inputs int32) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("amix=inputs=%d", inputs),
	}
}

func Loudnorm(i, tp int32) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("loudnorm=I=%d:TP=%d", i, tp),
	}
}

func ADelay(delays int32) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("adelay=delays=%ds:all=1", delays),
	}
}
