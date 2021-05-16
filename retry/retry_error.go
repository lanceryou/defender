package retry

// 重试错误接口
type RetryError interface {
	IsRetryError(err error) bool
}

type RetryErrorFunc func(err error) bool

func (e RetryErrorFunc) IsRetryError(err error) bool {
	return e(err)
}

func nopRetryError(err error) bool {
	return true
}

type RetryErrs struct {
	Errs []error
}

func (r *RetryErrs) IsRetryError(err error) bool {
	for _, e := range r.Errs {
		if e == err {
			return false
		}
	}

	return true
}
