package math

import "golang.org/x/exp/constraints"

// 接近n的最小偶数
// eg： CeilEven(41) is 40, CeilEven(30) is also 30.
func CeilEven[T constraints.Integer](n T) T {
	if n%2 == 0 {
		return n
	}
	return n - 1
}

type Rational[N, D constraints.Integer | constraints.Float] struct {
	Num N
	Den D
}

func Fraction[N, D constraints.Integer | constraints.Float](num N, den D) *Rational[N, D] {
	return &Rational[N, D]{
		Num: num,
		Den: den,
	}
}
