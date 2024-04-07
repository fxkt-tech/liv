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

// 音频倍速
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
		content: fmt.Sprintf("amix=inputs=%d:duration=first", inputs),
	}
}

func Loudnorm(i, tp int32) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("loudnorm=I=%d:TP=%d", i, tp),
	}
}

func ADelay(delays float32) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("adelay=delays=%fs:all=1", delays),
	}
}

func ANullSrc(duration float32) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("anullsrc=d=%f", duration),
	}
}
