package math

import "golang.org/x/exp/constraints"

// 内接
func Inscribed[T constraints.Integer](outW, outH, w, h T) (T, T) {
	if w*outH >= outW*h {
		return outW, outW * h / w
	} else {
		return outH, outH * w / h
	}
}

func Position[T constraints.Integer](outW, outH, w, h, xp, yp T) (T, T) {
	if w*outH >= outW*h {
		return 0, (outH - h) * yp / 100
	} else {
		return (outW - w) * xp / 100, 0
	}
}
