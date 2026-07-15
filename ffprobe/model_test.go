package ffprobe

import (
	"encoding/json"
	"errors"
	"testing"
	"time"
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

func TestProbeDurationPreservesLongMediaMicroseconds(t *testing.T) {
	raw := []byte(`{
  "streams": [{"codec_type": "video", "duration": "36000.123456"}],
  "format": {"duration": "36000.123456"}
}`)

	var got Probe
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	want := 10*time.Hour + 123456*time.Microsecond
	if got.Streams[0].Duration.Std() != want {
		t.Fatalf("stream duration = %s, want %s", got.Streams[0].Duration.Std(), want)
	}
	if got.Format.Duration.Std() != want {
		t.Fatalf("format duration = %s, want %s", got.Format.Duration.Std(), want)
	}
}

func TestProbeDurationJSONForms(t *testing.T) {
	tests := []string{`"1.000001"`, `1.000001`, `"N/A"`, `null`}
	wants := []time.Duration{time.Second + time.Microsecond, time.Second + time.Microsecond, 0, 0}

	for index, data := range tests {
		var got Duration
		if err := json.Unmarshal([]byte(data), &got); err != nil {
			t.Fatalf("json.Unmarshal(%s) error = %v", data, err)
		}
		if got.Std() != wants[index] {
			t.Fatalf("json.Unmarshal(%s) = %s, want %s", data, got.Std(), wants[index])
		}
	}
}

func TestProbeDurationRejectsInvalidValues(t *testing.T) {
	tests := []string{`"invalid"`, `"-1"`, `"9223372036854775808"`}
	for _, data := range tests {
		var got Duration
		if err := json.Unmarshal([]byte(data), &got); !errors.Is(err, ErrInvalidDuration) {
			t.Fatalf("json.Unmarshal(%s) error = %v, want ErrInvalidDuration", data, err)
		}
	}
}
