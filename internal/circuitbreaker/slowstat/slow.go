package slowstat

import (
	"sync/atomic"
	"time"

	"github.com/lanceryou/defender/internal/base"
	"github.com/lanceryou/defender/pkg/timering"
)

type SlowStat struct {
	slowResponseMs int64 // 慢回复值
	ratio          float64
	slowBuckets    []SlowBucket
	*timering.TimeRing
}

func NewSlowStat(slowResponseMs int64, ring *timering.TimeRing, ratio float64) *SlowStat {
	ss := &SlowStat{
		slowResponseMs: slowResponseMs,
		ratio:          ratio,
	}

	// ss.slowBuckets = make([]SlowBucket, intervalInMs/bucketCount)

	bucketResetArray := make([]timering.ResetBucket, len(ss.slowBuckets))
	for i := 0; i < len(bucketResetArray); i++ {
		bucketResetArray[i] = &ss.slowBuckets[i]
	}
	ss.TimeRing = ring
	ss.SetResetBuckets(bucketResetArray)
	return ss
}

func (s *SlowStat) MatchCount() int64 {
	var cnt int64
	for _, bucket := range s.slowBuckets {
		cnt += bucket.slowCount
	}

	return cnt
}

func (s *SlowStat) Total() int64 {
	var cnt int64
	for _, bucket := range s.slowBuckets {
		cnt += bucket.totalCount
	}

	return cnt
}

func (s *SlowStat) Stat(fn func() error, cr func(match bool, reach bool)) func() error {
	return func() error {
		start := time.Now().UnixNano()
		err := fn()
		end := time.Now().UnixNano()
		elapsed := end - start

		idx := s.CurrentIndex(start)
		match := elapsed >= s.slowResponseMs
		if match {
			atomic.AddInt64(&s.slowBuckets[idx].slowCount, 1)
		}
		atomic.AddInt64(&s.slowBuckets[idx].totalCount, 1)

		cr(match, s.ratioDetect())
		return err
	}
}

func (s *SlowStat) String() string {
	return "slowStat"
}

func (s *SlowStat) ratioDetect() bool {
	return base.FloatGte(float64(s.MatchCount())/float64(s.Total()), s.ratio)
}

//
// 慢回复统计
// 获取当前桶，在桶里统计计数
type SlowBucket struct {
	slowCount  int64 // 慢回复总数
	totalCount int64 // 请求数
}

func (s *SlowBucket) Reset() {
	atomic.StoreInt64(&s.slowCount, 0)
	atomic.StoreInt64(&s.totalCount, 0)
}
