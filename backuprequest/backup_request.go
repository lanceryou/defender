package backuprequest

import (
	"context"
	"sync/atomic"
	"time"
	"unsafe"
)

type Backup struct {
	BackupRequestMs time.Duration
}

type BackupResult struct {
	Ret unsafe.Pointer
}

func (r BackupResult) Cas(new unsafe.Pointer) bool {
	return atomic.CompareAndSwapPointer(&r.Ret, r.Ret, new)
}

// 业务自己做资源保护
/*
 * var ret int
 * Backup{time.Second}.Execute(ctx, func() error{
 *        tmp, err := apply.exec()
 *        if err != nil{
 *              return err
 *        }
 *        if BackupResult{Ret:unsafe.Pointer(&ret)}.Cas(unsafe.Pointer(&tmp)){
 *                // has response, do not
 *                return err
 *        }
 *        return err
 * })
 */
func (b *Backup) Execute(ctx context.Context, fn func() error) error {
	errChan := make(chan error)
	async := func() { errChan <- fn() }
	errChanFn := func(total int) {
		var cnt int
		for range errChan {
			cnt++
			if cnt == total {
				return
			}
		}
	}

	go async()
	ticker := time.NewTicker(b.BackupRequestMs)
	hasBackRequest := false
	for {
		select {
		case <-ctx.Done():
			cnt := 1
			if hasBackRequest {
				cnt++
			}
			go errChanFn(cnt)
			return ctx.Err()
		case <-ticker.C:
			hasBackRequest = true
			go async() // 启动backup request
			ticker.Stop()
		case err := <-errChan:
			if hasBackRequest {
				go errChanFn(1)
			}
			return err
		}
	}
}
