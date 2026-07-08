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

func TestZScale(t *testing.T) {
	got := ZScale("pin=bt2020:tin=smpte2084:min=bt2020nc:rin=tv:t=linear:npl=100").String()

	if !strings.Contains(got, "zscale=pin=bt2020:tin=smpte2084:min=bt2020nc:rin=tv:t=linear:npl=100") {
		t.Fatalf("ZScale() = %q, want zscale filter", got)
	}
}

func TestTonemap(t *testing.T) {
	got := Tonemap("tonemap=mobius:param=0.3:desat=0").String()

	if !strings.Contains(got, "tonemap=tonemap=mobius:param=0.3:desat=0") {
		t.Fatalf("Tonemap() = %q, want tonemap filter", got)
	}
}

func TestFormat(t *testing.T) {
	got := Format("gbrpf32le").String()

	if !strings.Contains(got, "format=gbrpf32le") {
		t.Fatalf("Format() = %q, want format filter", got)
	}
}

func TestSetSAR(t *testing.T) {
	got := SetSAR("1").String()

	if !strings.Contains(got, "setsar=1") {
		t.Fatalf("SetSAR() = %q, want setsar filter", got)
	}
}
