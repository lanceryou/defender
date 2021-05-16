package retry

import (
	"time"
)

// 重试组件
// backoff 策略（线性退避，随机退避）
// 不可重试策略
// 总时长
// 总次数
type Retryer struct {
	opt Options
}

// 重试
func (r *Retryer) Retry(fn func() error) (err error) {
	now := time.Now().UnixNano()
	for i := 0; i <= r.opt.maxCount; i++ {
		if err = fn(); err == nil {
			return nil
		}
		// 不可重试错误直接返回
		if !r.opt.ref.IsRetryError(err) {
			return err
		}

		elapsed := now - time.Now().UnixNano()
		if err = r.opt.bo.Wait(r.opt.maxDelay - elapsed); err != nil {
			return err
		}
	}
	return
}

func NewRetryer(opts ...Option) *Retryer {
	opt := Options{
		ref:      RetryErrorFunc(nopRetryError),
		bo:       BackoffFunc(NopBackoff),
		maxCount: 1,
		maxDelay: int64(time.Second),
	}
	for _, o := range opts {
		o(&opt)
	}

	return &Retryer{
		opt: opt,
	}
}
