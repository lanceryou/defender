package backuprequest

import (
	"context"
	"time"
)

type Backup struct {
	BackupRequestMs time.Duration
}

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
