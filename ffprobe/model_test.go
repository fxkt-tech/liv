package ffprobe

import (
	"encoding/json"
	"testing"
)

func TestProbeStreamUnmarshalColorFields(t *testing.T) {
	raw := []byte(`{
  "streams": [
    {
      "codec_type": "video",
      "color_space": "bt2020nc",
      "color_transfer": "smpte2084",
      "color_primaries": "bt2020"
    }
  ]
}`)

	var got Probe
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if len(got.Streams) != 1 {
		t.Fatalf("len(streams) = %d, want 1", len(got.Streams))
	}
	stream := got.Streams[0]
	if stream.ColorSpace != "bt2020nc" {
		t.Fatalf("ColorSpace = %q, want bt2020nc", stream.ColorSpace)
	}
	if stream.ColorTransfer != "smpte2084" {
		t.Fatalf("ColorTransfer = %q, want smpte2084", stream.ColorTransfer)
	}
	if stream.ColorPrimaries != "bt2020" {
		t.Fatalf("ColorPrimaries = %q, want bt2020", stream.ColorPrimaries)
	}
}
