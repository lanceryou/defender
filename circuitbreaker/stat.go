package circuitbreaker

// 统计熔断信息
type CircuitBreakerStat interface {
	Stat(fn func() error) func() error
	Total() int64
	MatchCount() int64
	String() string
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
