package base

import (
	"time"
)

// UnixMs
func UnixMs(t time.Time) int64 {
	mills := time.Millisecond / time.Nanosecond
	return t.UnixNano() / int64(mills)
}
