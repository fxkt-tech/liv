package filter

import (
	"strings"
	"testing"
)

func TestLut3D(t *testing.T) {
	got := Lut3D("/tmp/filter.cube").String()

	if !strings.Contains(got, "lut3d=file=/tmp/filter.cube:interp=tetrahedral") {
		t.Fatalf("Lut3D() = %q, want lut3d filter", got)
	}
}
