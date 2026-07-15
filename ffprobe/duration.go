package ffprobe

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"
)

var ErrInvalidDuration = errors.New("invalid ffprobe duration")

// Duration preserves ffprobe decimal-second values at time.Duration precision.
type Duration time.Duration

func ParseDuration(value string) (Duration, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "N/A" {
		return 0, nil
	}

	seconds, ok := new(big.Rat).SetString(value)
	if !ok || seconds.Sign() < 0 {
		return 0, fmt.Errorf("%w: %q", ErrInvalidDuration, value)
	}

	nanoseconds := new(big.Rat).Mul(seconds, big.NewRat(int64(time.Second), 1))
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(nanoseconds.Num(), nanoseconds.Denom(), remainder)
	if new(big.Int).Lsh(remainder, 1).Cmp(nanoseconds.Denom()) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if !quotient.IsInt64() {
		return 0, fmt.Errorf("%w: %q overflows time.Duration", ErrInvalidDuration, value)
	}
	return Duration(time.Duration(quotient.Int64())), nil
}

func (d Duration) Std() time.Duration {
	return time.Duration(d)
}

func (d Duration) Seconds() float64 {
	return time.Duration(d).Seconds()
}

func (d Duration) String() string {
	return time.Duration(d).String()
}

func (d Duration) MarshalJSON() ([]byte, error) {
	if d < 0 {
		return nil, fmt.Errorf("%w: %s", ErrInvalidDuration, time.Duration(d))
	}
	nanoseconds := int64(d)
	seconds := nanoseconds / int64(time.Second)
	fraction := nanoseconds % int64(time.Second)
	if fraction == 0 {
		return json.Marshal(strconv.FormatInt(seconds, 10))
	}
	value := fmt.Sprintf("%d.%09d", seconds, fraction)
	return json.Marshal(strings.TrimRight(value, "0"))
}

func (d *Duration) UnmarshalJSON(data []byte) error {
	if d == nil {
		return fmt.Errorf("%w: nil receiver", ErrInvalidDuration)
	}
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		*d = 0
		return nil
	}

	var value string
	if len(data) > 0 && data[0] == '"' {
		if err := json.Unmarshal(data, &value); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidDuration, err)
		}
	} else {
		value = string(data)
	}

	parsed, err := ParseDuration(value)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
