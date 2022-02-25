package circuitbreaker

import (
	"github.com/lanceryou/defender/internal/circuitbreaker/errorstat"
	"github.com/lanceryou/defender/internal/circuitbreaker/slowstat"
	"github.com/lanceryou/defender/pkg/timering"
)

// 统计熔断信息
type CircuitBreakerStat interface {
	Stat(fn func() error, cr func(match bool, reach bool)) func() error
	String() string
}

func NewSlowStat(slowResponseMs int64, ring *timering.TimeRing, ratio float64) CircuitBreakerStat {
	return slowstat.NewSlowStat(slowResponseMs, ring, ratio)
}

func NewErrStat(ring *timering.TimeRing, ratio float64, total int64) CircuitBreakerStat {
	return errorstat.NewErrorStat(ring, ratio, total)
}

var (
	statsMap = make(map[string]CircuitBreakerStat)
)

func RegisterStat(stat CircuitBreakerStat) {
	statsMap[stat.String()] = stat
}

func UnRegisterStat(name string) {
	delete(statsMap, name)
}

func GetStat(name string) CircuitBreakerStat {
	return statsMap[name]
}
