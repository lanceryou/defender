package backuprequest

import (
	"context"
	"time"
)

func Do(ctx context.Context, brMs time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	bk := NewBackup(&timer{})
	return bk.Do(ctx, brMs, fn)
}

// BackupTimer backup timer interface
type BackupTimer interface {
	After(time.Duration) <-chan struct{}
	Stop()
}

type timer struct {
	tm *time.Timer
	C  chan struct{}
}

func (t *timer) After(out time.Duration) <-chan struct{} {
	t.tm = time.NewTimer(out)
	t.C = make(chan struct{}, 1)
	go func() {
		<-t.tm.C
		close(t.C)
	}()
	return t.C
}

func (t *timer) Stop() {
	t.tm.Stop()
}

func NewBackup(bt BackupTimer) *Backup {
	return &Backup{
		t: bt,
	}
}

type Backup struct {
	t BackupTimer
}

/*
 * ret, err := Backup{time.Second}.Do(ctx, func() (interface{}, error){
 *        return apply.exec()
 * })
 */
func (b *Backup) Do(ctx context.Context, brMs time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	retChan := make(chan interface{})
	async := func() {
		ret, err := fn()
		if err != nil {
			retChan <- err
		}
		retChan <- ret
	}
	retChanFn := func(total int) {
		var cnt int
		for range retChan {
			cnt++
			if cnt == total {
				return
			}
		}
	}

	go async()
	hasBackRequest := false
	for {
		select {
		case <-ctx.Done():
			b.t.Stop()
			cnt := 1
			if hasBackRequest {
				cnt++
			}
			go retChanFn(cnt)
			return nil, ctx.Err()
		case <-b.t.After(brMs):
			hasBackRequest = true
			go async() // start backup request
			b.t.Stop()
		case ret := <-retChan:
			b.t.Stop()
			if hasBackRequest {
				go retChanFn(1)
			}

			if err, ok := ret.(error); ok {
				return nil, err
			}
			return ret, nil
		}
	}
}
