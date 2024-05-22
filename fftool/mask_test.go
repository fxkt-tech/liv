package fftool_test

import (
	"testing"

	"github.com/fxkt-tech/liv/fftool"
)

func TestMask(t *testing.T) {
	err := fftool.GenMask(fftool.CoreAlphaLinear, 1080, 1920, "mask.png")
	if err != nil {
		t.Fatal(err)
	}
}
