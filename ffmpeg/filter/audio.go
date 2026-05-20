package filter

import (
	"fmt"

	"github.com/fxkt-tech/liv/fftool/naming"
	"golang.org/x/exp/constraints"
)

// 音频流复制成多份
func ASplit(n int, opts ...FilterOpt) *MultiFilter {
	return &MultiFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("asplit=%d", n), opts...),
		counts:  n,
	}
}

// 音频倍速
func ATempo[T constraints.Integer | constraints.Float | string](expr T, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("atempo=%v", expr), opts...),
	}
}

// 音频帧显示时间戳
func ASetPTS(expr string, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("asetpts=%s", expr), opts...),
	}
}

func AMix(inputs int32, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("amix=inputs=%d:duration=first:normalize=0", inputs), opts...),
	}
}

// 音频淡入淡出
func AFade(t string, st, d float32, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf(
			"afade=t=%s:st=%.2f:d=%.2f",
			t, st, d,
		), opts...),
	}
}

func Loudnorm(i, lra, tp float32, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("loudnorm=i=%f:lra=%f:tp=%f:print_format=json", i, lra, tp), opts...),
	}
}

func LoudnormDoublePass(i, lra, tp float32, mi, mlra, mtp, mthres float32, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name: naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf(
			"loudnorm=i=%f:lra=%f:tp=%f:measured_i=%f:measured_lra=%f:measured_tp=%f:measured_thresh=%f:print_format=json",
			i, lra, tp, mi, mlra, mtp, mthres,
		), opts...),
	}
}

func ADelay(delays float32, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("adelay=delays=%f:all=1", delays*1000), opts...),
	}
}

func ANullSrc(duration float32, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("anullsrc=r=44100:d=%fs", duration), opts...),
	}
}

// 音频循环
// loop: 循环次数，-1表示无限循环，0表示不循环
// size: 每次循环的最大采样数，默认0表示整个输入
func ALoop(loop int, size int64, opts ...FilterOpt) *SingleFilter {
	return &SingleFilter{
		name:    naming.Default.Gen(),
		content: joinFilter(fmt.Sprintf("aloop=loop=%d:size=%d", loop, size), opts...),
	}
}
