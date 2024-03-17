package util

import "golang.org/x/exp/constraints"

func FixPixelLen[I constraints.Signed](l I) I {
	if l == 0 {
		return -2
	}
	return l
}
