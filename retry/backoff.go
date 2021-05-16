package retry

import (
	"errors"
	"time"
)

var (
	TimeoutErr = errors.New("backoff timeout err")
)

type Backoff interface {
	Wait(ttl int64) error
}

type BackoffFunc func(ttl int64) error

func (b BackoffFunc) Wait(ttl int64) error {
	return b(ttl)
}

func NopBackoff(ttl int64) error {
	if ttl < 0 {
		return TimeoutErr
	}

	return nil
}

type LinearBackoff struct {
	delay int64
}

func (b *LinearBackoff) Wait(ttl int64) error {
	if ttl < 0 || ttl-b.delay < 0 {
		return TimeoutErr
	}

	time.Sleep(time.Duration(b.delay))
	return nil
}
