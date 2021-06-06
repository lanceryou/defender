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

type StatDetector interface {
	Detect(match int64, total int64) bool
}

type StatDetectFunc func(match int64, total int64) bool

func (f StatDetectFunc) Detect(match int64, total int64) bool {
	return f(match, total)
}

func RatioDetect(ratio float64) StatDetectFunc {
	return func(match int64, total int64) bool {
		return base.FloatGte(float64(match)/float64(total), ratio)
	}
}

func TotalDetect(total int64) StatDetectFunc {
	return func(match int64, total int64) bool {
		return match >= total
	}
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
			if !c.reachRetryTimestamp() {
				return CircuitBreakerOpenErr
			}
			// if cas fail,it means state has change,so we need load again
			if !state.cas(Open, HalfOpen) {
				continue
			}
		}
		// half open 尝试probe
		err := c.stat(fn)
		c.tryUpdateState(state, err)
		return err
	}
}

func (c *CircuitBreaker) reachRetryTimestamp() bool {
	return time.Now().UnixNano() >= c.nextRetryTimestampMs
}

func (c *CircuitBreaker) stat(fn func() error) error {
	return c.opt.stat.Stat(fn)()
}

func (c *CircuitBreaker) updateNextRetryTimestampMs() {
	mills := time.Millisecond / time.Nanosecond
	currentTimeMills := time.Now().UnixNano() / int64(mills)
	atomic.StoreInt64(&c.nextRetryTimestampMs, currentTimeMills)
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
			c.updateNextRetryTimestampMs()
			c.state.Store(Open)
		}
		return
	}

	// 当前状态closed，err != nil
	if c.opt.detect.Detect(c.opt.stat.MatchCount(), c.opt.stat.Total()) {
		c.state.Store(Open)
		c.updateNextRetryTimestampMs()
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
