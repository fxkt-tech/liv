package ffprobe_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/fxkt-tech/liv/ffprobe"
)

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
