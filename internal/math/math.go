package math

import "golang.org/x/exp/constraints"

// 最接近n的最小偶数
// eg： CeilOdd(41) is 40, CeilOdd(30) is also 30.
func CeilOdd[T constraints.Integer](n T) T {
	if n%2 == 0 {
		return n
	}
	return n - 1
}

func CeilOddInt32(n int32) int32 {
	if n%2 == 0 {
		return n
	}
	return n - 1
}
