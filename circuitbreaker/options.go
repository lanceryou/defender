package circuitbreaker

// 拦截器设计
// metrics 资源设计
type Options struct {
	stat           CircuitBreakerStat
	detect         StatDetector
	retryTimeoutMs uint64
}

type Option func(*Options)

func WithCircuitBreakerStat(name string) Option {
	return func(options *Options) {
		options.stat = GetStat(name)
	}
}

func WithDetectName(name string) Option {
	return func(options *Options) {
		options.detect = nil
	}
}

func WithRetryTimeoutMs(retryTimeoutMs uint64) Option {
	return func(options *Options) {
		options.retryTimeoutMs = retryTimeoutMs
	}
}
