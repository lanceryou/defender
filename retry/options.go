package retry

type Options struct {
	rctx     RetryContext
	bo       Backoff
	maxDelay int64
	maxCount int
}

type Option func(*Options)

func WithMaxDelay(maxdelay int64) Option {
	return func(o *Options) {
		o.maxDelay = maxdelay
	}
}

func WithMaxCount(cnt int) Option {
	return func(o *Options) {
		o.maxCount = cnt
	}
}

func WithBackoff(b Backoff) Option {
	return func(o *Options) {
		o.bo = b
	}
}

func WithRetryContext(r RetryContext) Option {
	return func(o *Options) {
		o.rctx = r
	}
}
