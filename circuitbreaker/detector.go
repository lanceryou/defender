package circuitbreaker

import (
	"github.com/lanceryou/defender/internal/base"
)

type CircuitDetector interface {
	Detect(stat string, match int64, total int64) bool
}

type CircuitDetectFunc func(stat string, match int64, total int64) bool

func (f CircuitDetectFunc) Detect(stat string, match int64, total int64) bool {
	return f(stat, match, total)
}

func RatioDetect(detect string, ratio float64) CircuitDetectFunc {
	return func(stat string, match int64, total int64) bool {
		return detect == stat && base.FloatGte(float64(match)/float64(total), ratio)
	}
}

func TotalDetect(detect string, max int64) CircuitDetectFunc {
	return func(stat string, match int64, total int64) bool {
		return detect == stat && match >= max
	}
}
