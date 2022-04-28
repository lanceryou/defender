package circuitbreaker

// 拦截器设计
// metrics 资源设计
type Options struct {
	stats          []CircuitBreakerStat
	retryTimeoutMs int64
}

type Option func(*Options)

func WithCircuitBreakerStat(stat ...CircuitBreakerStat) Option {
	return func(options *Options) {
		options.stats = stat
	}
}

func WithRetryTimeoutMs(retryTimeoutMs int64) Option {
	return func(options *Options) {
		options.retryTimeoutMs = retryTimeoutMs
	}
}
