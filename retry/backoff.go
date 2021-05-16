package retry

import (
	"errors"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

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
	Delay int64
}

func (b *LinearBackoff) Wait(ttl int64) error {
	if ttl < 0 || ttl-b.Delay < 0 {
		return TimeoutErr
	}

	time.Sleep(time.Duration(b.Delay))
	return nil
}

type RandomBackoff struct {
	Jitter time.Duration
	Delay  time.Duration
}

func (b *RandomBackoff) Wait(ttl int64) error {
	if ttl < 0 {
		return TimeoutErr
	}

	jitter := rand.Int63n(int64(b.Jitter))
	delay := int64(b.Delay) + jitter
	if ttl < delay {
		return TimeoutErr
	}

	time.Sleep(time.Duration(jitter))
	return nil
}
