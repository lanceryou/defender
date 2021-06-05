package base

import (
	"math"
)

const precision = 0.00000001

func Gte(l, r float64) bool {
	return l > r || Eq(l, r)
}

func Eq(l, r float64) bool {
	return math.Abs(l-r) < precision
}
