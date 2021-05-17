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

	go async()
	ticker := time.NewTicker(b.BackupRequestMs)
	hasBackRequest := false
	for {
		select {
		case <-ctx.Done():
			if hasBackRequest {
				go func() { <-errChan }()
			}
			return ctx.Err()
		case <-ticker.C:
			hasBackRequest = true
			go async() // 启动backup request
			ticker.Stop()
		case err := <-errChan:
			if hasBackRequest {
				go func() { <-errChan }()
			}
			return err
		}
	}
}
