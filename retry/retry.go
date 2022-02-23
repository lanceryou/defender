package retry

import (
	"context"
	"time"
)

// 重试组件
// backoff 策略（线性退避，随机退避）
// 不可重试策略
// 总时长
// 总次数
type Retrier struct {
	opt Options
}

// 重试
// 限制单点重试
// 超时处理
func (r *Retrier) Retry(ctx context.Context, fn func() error) (err error) {
	now := time.Now().UnixNano()
	for i := 0; i <= r.opt.maxCount; i++ {
		if err = fn(); err == nil {
			return nil
		}
		// 不可重试直接返回错误
		if !r.opt.rctx.CanRetry(ctx) {
			return err
		}

		elapsed := now - time.Now().UnixNano()
		if err = r.opt.bo.Wait(r.opt.maxDelay - elapsed); err != nil {
			return err
		}
	}
	return
}

func NewRetryer(opts ...Option) *Retrier {
	opt := Options{
		bo:       BackoffFunc(NopBackoff),
		rctx:     RetryContextFunc(NopCanRetry),
		maxCount: 1,
		maxDelay: int64(time.Second),
	}
	for _, o := range opts {
		o(&opt)
	}

	return &Retrier{
		opt: opt,
	}
}
