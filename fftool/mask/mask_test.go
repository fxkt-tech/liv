package mask_test

import (
	"testing"

	"github.com/fxkt-tech/liv/fftool/mask"
)

func TestMask(t *testing.T) {
	err := mask.Gen(mask.CoreAlphaLinear, 1080, 1920, "mask.png")
	if err != nil {
		t.Fatal(err)
	}
}
