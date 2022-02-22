package backuprequest

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestBackup_ExecuteWithResult(t *testing.T) {
	ts := []struct {
		brMS   time.Duration
		err    error
		result int64
		cnt    int64
	}{
		{
			brMS: time.Millisecond * 10,
			err:  errors.New("what error"),
		},
		{
			brMS:   time.Millisecond * 10,
			result: 1,
		},
		{
			brMS:   time.Millisecond * 10,
			result: 1,
			cnt:    1,
		},
		{
			brMS:   time.Millisecond * 10,
			result: 1,
			cnt:    2,
		},
	}

	for _, s := range ts {
		now := time.Now()
		cnt := s.cnt
		if cnt > 1 {
			cnt = 1
		}
		r, err := Do(context.Background(), s.brMS, func() (interface{}, error) {
			cnt := atomic.AddInt64(&s.cnt, -1)
			if cnt >= 0 {
				time.Sleep(s.brMS + time.Microsecond*10)
			}
			if s.err != nil {
				return nil, s.err
			}

			return s.result, nil
		})

		since := time.Since(now).Milliseconds()
		expect := cnt * (s.brMS + time.Microsecond*10).Milliseconds()
		if since != expect {
			t.Errorf("expect cost %v, but %v", expect, since)
		}

		if err != nil && err.Error() != s.err.Error() {
			t.Errorf("expect err %v, but %v", s.err, err)
		}

		if r != nil && r.(int64) != s.result {
			t.Errorf("expect result %v, but %v", s.result, r.(int64))
		}
	}
}

func BenchmarkBackup_ExecuteWithResult(b *testing.B) {
	back := Backup{t: &timer{}}
	for n := 0; n < b.N; n++ {
		back.Do(context.TODO(), time.Millisecond*10, func() (interface{}, error) {
			return 1, nil
		})
	}
}
