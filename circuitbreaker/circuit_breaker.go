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

// slowRT
// err
// metrics
// 一个资源一个CircuitBreaker？
// metrics 信息收集？
type CircuitBreaker struct {
	opt                  Options
	state                State
	nextRetryTimestampMs int64
}

/* Allow
 * 全部状态转移放这里
 *
 */
func (c *CircuitBreaker) Allow(fn func() error) error {
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
		err := c.stat(fn)
		c.tryUpdateState(state, err)
		return err
	}
}

func (c *CircuitBreaker) reachRetryTimestamp(t time.Time) bool {
	return base.UnixMs(t) >= atomic.LoadInt64(&c.nextRetryTimestampMs)
}

func (c *CircuitBreaker) stat(fn func() error) error {
	sf := fn
	for _, stat := range c.opt.stats {
		sf = stat.Stat(sf)
	}

	return sf()
}

func (c *CircuitBreaker) updateNextRetryTimestampMs(t time.Time) {
	currentTimeMills := base.UnixMs(t)
	atomic.StoreInt64(&c.nextRetryTimestampMs, currentTimeMills+c.opt.retryTimeoutMs)
	return
}

func (c *CircuitBreaker) tryUpdateState(cur State, err error) {
	if cur == Open {
		return
	}

	// 	状态不需要扭转
	if cur == Closed && err == nil {
		return
	}
	// 当前状态是half open probe成功转closed，失败open
	if cur == HalfOpen {
		if err == nil {
			c.state.Store(Closed)
		} else {
			c.updateNextRetryTimestampMs(time.Now())
			c.state.Store(Open)
		}
		return
	}

	for _, detector := range c.opt.detects {
		for _, stat := range c.opt.stats {
			if detector.Detect(stat.String(), stat.MatchCount(), stat.Total()) {
				c.state.Store(Open)
				c.updateNextRetryTimestampMs(time.Now())
			}
		}
	}
}

func NewCircuitBreaker(option ...Option) *CircuitBreaker {
	var opt Options
	for _, o := range option {
		o(&opt)
	}

	return &CircuitBreaker{
		opt:   opt,
		state: Closed,
	}
}
