package math

import (
	"testing"
)

func TestCeilOdd(t *testing.T) {
	var n1 int64 = 10
	x1 := CeilOddInt64(n1)
	if x1 != 10 {
		t.Error(x1)
	}

	var n2 int64 = 12
	x2 := CeilOddInt64(n2)
	if x2 != 10 {
		t.Error(x2)
	}
}
