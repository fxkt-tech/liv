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
		content: fmt.Sprintf("amix=inputs=%d:duration=first:normalize=0", inputs),
	}
}

// 音频淡入淡出
func AFade(t string, st, d float32) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: fmt.Sprintf(
			"afade=t=%s:st=%.2f:d=%.2f",
			t, st, d,
		),
	}
}

func Loudnorm(i, lra, tp float32) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("loudnorm=i=%f:lra=%f:tp=%f:print_format=json", i, lra, tp),
	}
}

func LoudnormDoublePass(i, lra, tp float32, mi, mlra, mtp, mthres float32) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: fmt.Sprintf(
			"loudnorm=i=%f:lra=%f:tp=%f:measured_i=%f:measured_lra=%f:measured_tp=%f:measured_thresh=%f:print_format=json",
			i, lra, tp, mi, mlra, mtp, mthres,
		),
	}
}

func ADelay(delays float32) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("adelay=delays=%f:all=1", delays*1000),
	}
}

func ANullSrc(duration float32) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("anullsrc=r=44100:d=%fs", duration),
	}
}
