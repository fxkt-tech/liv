package math

import (
	"testing"
)

func TestCeilEven(t *testing.T) {
	var n1 int64 = 10
	x1 := CeilEven(n1)
	if x1 != 10 {
		t.Error(x1)
	}

	var n2 int64 = 11
	x2 := CeilEven(n2)
	if x2 != 10 {
		t.Error(x2)
	}
}
