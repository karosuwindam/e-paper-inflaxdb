package controller

import (
	"context"
	"errors"
	"log/slog"
	"time"
)

var (
	shutdown chan struct{}
	done     chan struct{}
)

func Init() error {
	shutdown = make(chan struct{}, 1)
	done = make(chan struct{}, 1)
	return nil
}

func Run(ctx context.Context) error {
	var oneshot chan bool = make(chan bool, 1)
	oneshot <- true
	slog.InfoContext(ctx, "Controller Run")
loop:
	for {
		select {
		case <-oneshot:
			slog.DebugContext(ctx, "onshot run ePaperUpdate")
			//e-paperの更新
			if err := ePaperUpdate(ctx); err != nil {
				slog.ErrorContext(ctx, "Failed to update e-paper", "error", err.Error())
			}
		case <-shutdown:
			slog.DebugContext(ctx, "run shutdown")
			done <- struct{}{}
			break loop
		case <-time.After(time.Minute * 5): // 5 minutes
			slog.DebugContext(ctx, "loop time after 5 run ePaperUpdate")
			//e-paperの更新
			if err := ePaperUpdate(ctx); err != nil {
				slog.ErrorContext(ctx, "Failed to update e-paper", "error", err.Error())
			}
		case <-ctx.Done():
			slog.DebugContext(ctx, "run ctx Done()")
			close(shutdown)
			break loop
		}
	}
	slog.InfoContext(ctx, "Controller Shutdown")
	return nil
}

func Stop(ctx context.Context) error {
	if len(shutdown) != 0 {
		return nil
	}
	slog.InfoContext(ctx, "Stop Wait")
	shutdown <- struct{}{}
	select {
	case <-done:
		break
	case <-time.After(time.Millisecond * 500):
		return errors.New("Stop timeout")
	}
	return nil
}
