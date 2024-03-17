package ffprobe_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/fxkt-tech/liv/ffprobe"
)

func TestExtract(t *testing.T) {
	var (
		ctx = context.Background()

		infile = "in.mp4"
	)

	fp, err := ffprobe.New().Input(infile).Extract(ctx)
	if err != nil {
		t.Fatal(err)
	}

	vStream := fp.GetFirstVideoStream()
	if vStream == nil {
		t.Fatal("file has no v stream")
	}

	t.Log(vStream)
}

func TestProbeRunRaw(t *testing.T) {
	ctx := context.Background()

	path := "/Users/justyer/Downloads/shot.mp4"
	probe := ffprobe.New().Input(path)
	raw, err := probe.RunRetRaw(ctx)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(raw))
	fmt.Println(strings.Contains(string(raw), `"rotation"`))
}
