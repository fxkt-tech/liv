package math

import "golang.org/x/exp/constraints"

// 整型绝对值
func Abs[T constraints.Integer | constraints.Float](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// 裁剪
func Clip[T constraints.Ordered](x, a, b T) T {
	return min(max(x, a), b)
}
