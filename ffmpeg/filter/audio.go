package filter

import (
	"fmt"

	"github.com/fxkt-tech/liv/internal/naming"
)

// 音频流复制成多份
func ASplit(n int) Filter {
	return &multiple{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("asplit=%d", n),
		counts:  n,
	}
}

// 音频帧显示时间戳
func ASetPTS(expr string) Filter {
	return &single{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("asetpts=%s", expr),
	}
}

func AMix(inputs int32) Filter {
	return &single{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("amix=inputs=%d", inputs),
	}
}

func Loudnorm(i, tp int32) Filter {
	return &single{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("loudnorm=I=%d:TP=%d", i, tp),
	}
}

func ADelay(delays int32) Filter {
	return &single{
		name:    naming.Default.Gen(),
		content: fmt.Sprintf("adelay=delays=%ds:all=1", delays),
	}
}
