package circuitbreaker

import (
	"errors"
	"sync/atomic"
	"time"

	"github.com/lanceryou/defender/internal/base"
)

var (
	CircuitBreakerOpenErr = errors.New("circuit breaker open err")
)

// 熔断降级
//  Circuit Breaker State Machine:
//
//				+-----------------------------------------------------------------------+
//				|                                                                       |
//				|                                                                       v
//		+----------------+                   +----------------+      Probe      +----------------+
//		|                |                   |                |<----------------|                |
//		|                |   Probe succeed   |                |                 |                |
//		|     Closed     |<------------------|    HalfOpen    |                 |      Open      |
//		|                |                   |                |   Probe failed  |                |
//		|                |                   |                +---------------->|                |
//		+----------------+                   +----------------+                 +----------------+

type State int32

const (
	Closed State = iota
	HalfOpen
	Open
)

func (s State) Load() State {
	return State(atomic.LoadInt32((*int32)(&s)))
}

func (s State) Store(state State) {
	atomic.StoreInt32((*int32)(&s), int32(state))
}

func (s State) cas(expect State, update State) bool {
	return atomic.CompareAndSwapInt32((*int32)(&s), int32(expect), int32(update))
}

type circuitBreaker struct {
	state                State
	nextRetryTimestampMs int64
	stat                 CircuitBreakerStat
	retryTimeoutMs       int64
}

func newCircuitBreaker(stat CircuitBreakerStat, rt int64) *circuitBreaker {
	return &circuitBreaker{
		state:          Closed,
		stat:           stat,
		retryTimeoutMs: rt,
	}
}

// slowRT
// err
// metrics
// 一个资源一个CircuitBreaker？
// metrics 信息收集？
type CircuitBreaker struct {
	opt    Options
	cbList []*circuitBreaker
}

func (c *CircuitBreaker) Allow(fn func() error) error {
	for _, cb := range c.cbList {
		if err := cb.Allow(fn); err != nil {
			return err
		}
	}

	return nil
}

/* Allow
 * 全部状态转移放这里
 *
 */
func (c *circuitBreaker) Allow(fn func() error) error {
	for {
		state := c.state.Load()
		if state == Open {
			// state open and no reach retry time. refuse request.
			if !c.reachRetryTimestamp(time.Now()) {
				return CircuitBreakerOpenErr
			}
			// if cas fail,it means state has change,so we need load again
			if !state.cas(Open, HalfOpen) {
				continue
			}
		}
		// half open try probe
		err := c.stat.Stat(fn, c.tryUpdateState)()
		return err
	}
}

func (c *circuitBreaker) reachRetryTimestamp(t time.Time) bool {
	return base.UnixMs(t) >= atomic.LoadInt64(&c.nextRetryTimestampMs)
}

func (c *circuitBreaker) updateNextRetryTimestampMs(t time.Time) {
	currentTimeMills := base.UnixMs(t)
	atomic.StoreInt64(&c.nextRetryTimestampMs, currentTimeMills+c.retryTimeoutMs)
	return
}

func (c *circuitBreaker) tryUpdateState(match bool, reach bool) {
	cur := c.state.Load()
	if cur == Open {
		return
	} else if cur == HalfOpen {
		if !match {
			c.state.Store(Closed)
		} else {
			// probe failed
			// HalfOpen to Open
			c.updateNextRetryTimestampMs(time.Now())
			c.state.Store(Open)
		}
	} else if cur == Closed && reach {
		//  Closed to Open
		c.state.Store(Open)
		c.updateNextRetryTimestampMs(time.Now())
	}
}

func NewCircuitBreaker(option ...Option) *CircuitBreaker {
	var opt Options
	for _, o := range option {
		o(&opt)
	}

	cb := &CircuitBreaker{
		opt: opt,
	}

	for _, stat := range opt.stats {
		cb.cbList = append(cb.cbList, newCircuitBreaker(stat, opt.retryTimeoutMs))
	}

	return cb
}
