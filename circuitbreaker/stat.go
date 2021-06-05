package circuitbreaker

import (
	"github.com/lanceryou/defender/internal/circuitbreaker/errorstat"
	"github.com/lanceryou/defender/internal/circuitbreaker/slowstat"
)

// 统计熔断信息
type CircuitBreakerStat interface {
	Stat(fn func() error) func() error
	Total() int64
	MatchCount() int64
	String() string
}

func NewSlowStat(slowResponseMs int64, intervalInMs uint32, bucketCount uint32) CircuitBreakerStat {
	return slowstat.NewSlowStat(slowResponseMs, intervalInMs, bucketCount)
}

func NewErrStat(intervalInMs uint32, bucketCount uint32) CircuitBreakerStat {
	return errorstat.NewErrorStat(intervalInMs, bucketCount)
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
