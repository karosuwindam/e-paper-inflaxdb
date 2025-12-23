package controller

import (
	"context"
	"epaperifdb/controller/button"
	"log/slog"
	"sync"
	"time"
)

func Button_Run(ctx context.Context) error {
	key := button.InitKey()
	slog.DebugContext(ctx, "e-Paper Button Start")
	var wg sync.WaitGroup
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
	loop:
		for {
			select {
			case <-shutdown_button:
				slog.DebugContext(ctx, "Button Run shutdown")
				break loop
			case <-ctx.Done():
				break loop
			case <-time.After(time.Microsecond + 10):
				key.KeyRead()
				for i := 0; i < 4; i++ {
					if key.Read(i) {
						slog.DebugContext(ctx, "Read Key Input")
						re_new <- true
						key.KeyInitNum(i)
					}
				}
			}
		}
	}(ctx)
	wg.Wait()
	slog.DebugContext(ctx, "e-Paper Button Stop")
	return nil
}

func Button_Stop() {
	shutdown_button <- struct{}{}
}
