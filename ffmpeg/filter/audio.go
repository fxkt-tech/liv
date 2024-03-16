package filter

import "fmt"

// 音频流复制成多份
func ASplit(name string, n int) Filter {
	return &multiple{
		name:    name,
		content: fmt.Sprintf("asplit=%d", n),
		counts:  n,
	}
}

func AMix(name string, inputs int32) Filter {
	return &single{
		name:    name,
		content: fmt.Sprintf("amix=inputs=%d", inputs),
	}
}

func Loudnorm(name string, i, tp int32) Filter {
	return &single{
		name:    name,
		content: fmt.Sprintf("loudnorm=I=%d:TP=%d", i, tp),
	}
}

func ADelay(name string, delays int32) Filter {
	return &single{
		name:    name,
		content: fmt.Sprintf("adelay=delays=%ds:all=1", delays),
	}
}
