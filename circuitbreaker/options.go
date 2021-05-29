package circuitbreaker

// 拦截器设计
// metrics 资源设计
type Options struct {
	stat   CircuitBreakerStat
	detect StatDetector
}

type Option func(*Options)

func WithCircuitBreakerStat(names ...string) Option {
	return func(options *Options) {
		for _, name := range names {
			options.stat = GetStat(name)
		}
	}
}

func WithDetectName(name string) Option {
	return func(options *Options) {
		options.detect = nil
	}
}
