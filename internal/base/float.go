package base

import (
	"math"
)

const precision = 0.00000001

func FloatGte(l, r float64) bool {
	return l > r || FloatEq(l, r)
}

func FloatEq(l, r float64) bool {
	return math.Abs(l-r) < precision
}
