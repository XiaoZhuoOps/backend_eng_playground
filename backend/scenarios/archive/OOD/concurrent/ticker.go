package OOD

import (
	"context"
	"time"
)

func periodicTaskRunner(ctx context.Context, interval time.Duration, task func(ctx context.Context)) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			task(ctx)
		}
	}
}
