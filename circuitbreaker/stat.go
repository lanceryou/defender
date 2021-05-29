package circuitbreaker

type StatFunc func() error

// 统计熔断信息
type CircuitBreakerStat interface {
	Stat(fn StatFunc) StatFunc
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
