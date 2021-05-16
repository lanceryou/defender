package backuprequest

import (
	"context"
	"time"
)

type Backup struct {
	BackupRequestMs int64
}

func (b *Backup) Request(ctx context.Context, fn func() error) error {
	errChan := make(chan error)
	async := func() { errChan <- fn() }

	go async()
	ticker := time.NewTicker(time.Duration(b.BackupRequestMs))
	isBackRequest := false
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			isBackRequest = true
			go async() // 启动backup request
			ticker.Stop()
		case err := <-errChan:
			if isBackRequest {
				go func() { <-errChan }()
			}
			return err
		}
	}
}
