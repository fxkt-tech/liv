package ffcut

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"time"
)

// Duration stores protocol time as a Go time.Duration and serializes it as
// an integer number of microseconds.
type Duration time.Duration

// NewDuration converts a Go duration into protocol time. Project v2 only
// accepts microsecond precision so JSON round trips are lossless.
func NewDuration(value time.Duration) (Duration, error) {
	if value%time.Microsecond != 0 {
		return 0, fmt.Errorf("%w: %s is not aligned to a microsecond", ErrInvalidDuration, value)
	}
	return Duration(value), nil
}

// DurationFromMicroseconds constructs protocol time from its JSON unit.
func DurationFromMicroseconds(value int64) (Duration, error) {
	const unit = int64(time.Microsecond)
	if value > math.MaxInt64/unit || value < math.MinInt64/unit {
		return 0, fmt.Errorf("%w: %d microseconds overflows time.Duration", ErrInvalidDuration, value)
	}
	return Duration(time.Duration(value * unit)), nil
}

// Std returns the underlying Go duration.
func (d Duration) Std() time.Duration {
	return time.Duration(d)
}

// Microseconds returns the exact Project v2 JSON representation.
func (d Duration) Microseconds() (int64, error) {
	value := time.Duration(d)
	if value%time.Microsecond != 0 {
		return 0, fmt.Errorf("%w: %s is not aligned to a microsecond", ErrInvalidDuration, value)
	}
	return int64(value / time.Microsecond), nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	value, err := d.Microseconds()
	if err != nil {
		return nil, err
	}
	return []byte(strconv.FormatInt(value, 10)), nil
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	if d == nil {
		return fmt.Errorf("%w: nil duration receiver", ErrInvalidDuration)
	}

	var value int64
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("%w: expected integer microseconds: %v", ErrInvalidDuration, err)
	}
	parsed, err := DurationFromMicroseconds(value)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// TimeRange is an absolute start and duration on either a source or timeline.
type TimeRange struct {
	Start    Duration `json:"start"`
	Duration Duration `json:"duration"`
}

// End returns Start + Duration and reports overflow.
func (r TimeRange) End() (Duration, error) {
	start := int64(r.Start)
	duration := int64(r.Duration)
	if duration > 0 && start > math.MaxInt64-duration {
		return 0, fmt.Errorf("%w: time range end overflows", ErrInvalidDuration)
	}
	if duration < 0 && start < math.MinInt64-duration {
		return 0, fmt.Errorf("%w: time range end underflows", ErrInvalidDuration)
	}
	return Duration(start + duration), nil
}
