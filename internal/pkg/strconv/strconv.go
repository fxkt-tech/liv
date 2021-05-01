package strconv

import "strconv"

func ToInt32(s string) (int32, error) {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}

func ToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func ToInt32ByFloatString(s string) (int32, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return int32(f), nil
}

func ToInt64ByFloatString(s string) (int64, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return int64(f), nil
}
