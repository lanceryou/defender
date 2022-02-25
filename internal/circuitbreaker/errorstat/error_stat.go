package errorstat

import (
	"sync/atomic"
	"time"

	"github.com/lanceryou/defender/internal/base"
	"github.com/lanceryou/defender/pkg/timering"
)

type errorStat struct {
	errBuckets []errBucket
	ratio      float64
	match      int64
	*timering.TimeRing
}

func NewErrorStat(ring *timering.TimeRing, ratio float64, match int64) *errorStat {
	ss := &errorStat{
		ratio: ratio,
		match: match,
	}

	// ss.errBuckets = make([]errBucket, intervalInMs/bucketCount)
	bucketResetArray := make([]timering.ResetBucket, len(ss.errBuckets))
	for i := 0; i < len(bucketResetArray); i++ {
		bucketResetArray[i] = &ss.errBuckets[i]
	}
	ss.TimeRing = ring
	ss.SetResetBuckets(bucketResetArray)
	return ss
}

func (s *errorStat) MatchCount() int64 {
	var cnt int64
	for _, bucket := range s.errBuckets {
		cnt += bucket.errCount
	}

	return cnt
}

func (s *errorStat) Total() int64 {
	var cnt int64
	for _, bucket := range s.errBuckets {
		cnt += bucket.totalCount
	}

	return cnt
}

func (s *errorStat) Stat(fn func() error, cr func(match bool, reach bool)) func() error {
	return func() error {
		err := fn()
		idx := s.CurrentIndex(time.Now().UnixNano())
		match := err != nil
		if match {
			atomic.AddInt64(&s.errBuckets[idx].errCount, 1)
		}
		atomic.AddInt64(&s.errBuckets[idx].totalCount, 1)

		cr(match, s.reachCircuit())
		return err
	}
}

func (s *errorStat) String() string {
	return "errStat"
}

func (s *errorStat) reachCircuit() bool {
	matchCount := s.MatchCount()
	totalCount := s.Total()
	return matchCount >= s.match ||
		base.FloatGte(float64(matchCount)/float64(totalCount), s.ratio)
}

type errBucket struct {
	errCount   int64 // 错误总数
	totalCount int64 // 请求数
}

func (s *errBucket) Reset() {
	atomic.StoreInt64(&s.errCount, 0)
	atomic.StoreInt64(&s.totalCount, 0)
}
