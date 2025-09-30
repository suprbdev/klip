package timeutil

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// ParseTTL parses strings like "10m", "2h", "1d".
// It also accepts plain integer seconds (e.g. "30").
// Returns an error for malformed input or negative durations.
func ParseTTL(s string) (time.Duration, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 0, errors.New("empty duration")
	}
	// Support days suffix
	if strings.HasSuffix(s, "d") {
		num := strings.TrimSuffix(s, "d")
		val, err := strconv.Atoi(num)
		if err != nil {
			return 0, err
		}
		if val < 0 {
			return 0, errors.New("negative duration")
		}
		return time.Hour * 24 * time.Duration(val), nil
	}
	// Let Go parse the rest (supports h,m,s,ms etc.)
	dur, err := time.ParseDuration(s)
	if err != nil {
		return 0, err
	}
	if dur < 0 {
		return 0, errors.New("negative duration")
	}
	return dur, nil
}
