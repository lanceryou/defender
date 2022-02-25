package timering

import (
	"fmt"
	"runtime"
	"sync/atomic"
)

type ResetBucket interface {
	Reset()
}

type Bucket struct {
	startTime int64
	ResetBucket
}

// time ring，按照时间计算bucket
// 环产生回绕时候需要reset
type TimeRing struct {
	intervalInMs     uint32
	bucketCount      uint32 // bucket 数量
	bucketLengthInMs uint32
	buckets          []Bucket
}

func NewTimeRing(intervalInMs uint32, bucketCount uint32) *TimeRing {
	if intervalInMs%bucketCount != 0 {
		panic(fmt.Errorf("time ring intervalInMs must be divide bucketCount."))
	}

	r := &TimeRing{
		intervalInMs:     intervalInMs,
		bucketCount:      bucketCount,
		bucketLengthInMs: intervalInMs / bucketCount,
	}

	return r
}

// 10001 - 1
// calculate bucket index
func (r *TimeRing) calcBucketIndex(now int64) int64 {
	idx := now / int64(r.bucketLengthInMs)
	return idx % int64(r.bucketCount)
}

// SetResetBuckets
func (r *TimeRing) SetResetBuckets(buckets []ResetBucket) {
	r.buckets = make([]Bucket, r.bucketCount)
	for i := range r.buckets {
		r.buckets[i] = Bucket{
			ResetBucket: buckets[i],
		}
	}
}

// CurrentIndex get current bucket index
func (r *TimeRing) CurrentIndex(timeMills int64) int64 {
	idx := r.calcBucketIndex(timeMills)
	bucketStart := calculateStartTime(timeMills, r.bucketLengthInMs)

	for {
		bucket := r.buckets[idx]
		// first enter
		startTime := atomic.LoadInt64(&bucket.startTime)
		if startTime == 0 {
			if atomic.CompareAndSwapInt64(&bucket.startTime, startTime, bucketStart) {
				atomic.StoreInt64(&bucket.startTime, bucketStart)
				return idx
			} else {
				runtime.Gosched()
			}
		} else if startTime == bucketStart {
			return idx
		} else if startTime < bucketStart {
			// enter next cycle,so we need reset origin value
			// if cas fail means simultaneous phenomenon
			if atomic.CompareAndSwapInt64(&bucket.startTime, startTime, bucketStart) {
				atomic.StoreInt64(&bucket.startTime, bucketStart)
				bucket.Reset()
				return idx
			} else {
				runtime.Gosched()
			}
		} else {
			// something error, why happened
			panic(fmt.Errorf("startTime %v > bucketStart %v", startTime, bucketStart))
		}
	}
}

func calculateStartTime(now int64, bucketLengthInMs uint32) int64 {
	return now - (now % int64(bucketLengthInMs))
}
