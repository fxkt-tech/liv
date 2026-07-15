package ffcut

import (
	"encoding/json"
	"errors"
	"math"
	"testing"
	"time"
)

func TestDurationJSONUsesMicroseconds(t *testing.T) {
	got, err := NewDuration(1500 * time.Microsecond)
	if err != nil {
		t.Fatalf("NewDuration() error = %v", err)
	}

	data, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if string(data) != "1500" {
		t.Fatalf("json.Marshal() = %s, want 1500", data)
	}

	var decoded Duration
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	if decoded.Std() != 1500*time.Microsecond {
		t.Fatalf("decoded = %s, want 1.5ms", decoded.Std())
	}
}

func TestDurationRejectsSubMicrosecondPrecision(t *testing.T) {
	_, err := NewDuration(time.Microsecond + time.Nanosecond)
	if !errors.Is(err, ErrInvalidDuration) {
		t.Fatalf("NewDuration() error = %v, want ErrInvalidDuration", err)
	}
}

func TestDurationRejectsJSONOverflow(t *testing.T) {
	var got Duration
	err := json.Unmarshal([]byte("9223372036854775807"), &got)
	if !errors.Is(err, ErrInvalidDuration) {
		t.Fatalf("json.Unmarshal() error = %v, want ErrInvalidDuration", err)
	}
}

func TestDurationRejectsFractionalJSON(t *testing.T) {
	var got Duration
	err := json.Unmarshal([]byte("1.5"), &got)
	if !errors.Is(err, ErrInvalidDuration) {
		t.Fatalf("json.Unmarshal() error = %v, want ErrInvalidDuration", err)
	}
}

func TestTimeRangeEndRejectsOverflow(t *testing.T) {
	start := Duration((math.MaxInt64 / int64(time.Microsecond)) * int64(time.Microsecond))
	value := TimeRange{Start: start, Duration: duration(time.Microsecond)}
	if _, err := value.End(); !errors.Is(err, ErrInvalidDuration) {
		t.Fatalf("End() error = %v, want ErrInvalidDuration", err)
	}
}
