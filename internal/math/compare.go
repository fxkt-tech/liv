package math

import "golang.org/x/exp/constraints"

// 整型绝对值
func Abs[T constraints.Integer | constraints.Float](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// 最大值
func Max[T constraints.Ordered](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// 最小值
func Min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	}
	return b
}

// 裁剪
func Clip[T constraints.Ordered](x, a, b T) T {
	return Min(Max(x, a), b)
}
