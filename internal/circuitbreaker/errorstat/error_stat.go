package errorstat

import (
	"sync/atomic"
	"time"

	"github.com/lanceryou/defender/internal/base"
)

type errorStat struct {
	errBuckets []errBucket
	*base.TimeRing
}

func NewErrorStat(intervalInMs uint32, bucketCount uint32) *errorStat {
	ss := &errorStat{}

	ss.errBuckets = make([]errBucket, intervalInMs/bucketCount)

	bucketResetArray := make([]base.ResetBucket, len(ss.errBuckets))
	for i := 0; i < len(bucketResetArray); i++ {
		bucketResetArray[i] = &ss.errBuckets[i]
	}
	ss.TimeRing = base.NewTimeRing(intervalInMs, bucketCount, bucketResetArray)
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

func (s *errorStat) Stat(fn func() error) func() error {
	return func() error {
		err := fn()
		idx := s.CurrentIndex(time.Now().UnixNano())
		if err != nil {
			atomic.AddInt64(&s.errBuckets[idx].errCount, 1)
		}
		atomic.AddInt64(&s.errBuckets[idx].totalCount, 1)
		return err
	}
}

func (s *errorStat) String() string {
	return "errStat"
}

type errBucket struct {
	errCount   int64 // 错误总数
	totalCount int64 // 请求数
}

func (s *errBucket) Reset() {
	atomic.StoreInt64(&s.errCount, 0)
	atomic.StoreInt64(&s.totalCount, 0)
}
