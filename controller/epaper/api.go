package epaper

import (
	"context"
	"epaperifdb/config"
	"log/slog"
)

type api struct{}

func Init() (*api, error) {
	if err := initEpaper(); err != nil {
		return nil, err
	}
	return &api{}, nil
}

func (a *api) TextPut(ctx context.Context, x, y int, text []string, size float64) {
	testPut(ctx, x, y, text, size)
}

func (a *api) ClearScreen(ctx context.Context) {
	ctx, span := config.TracerS(ctx, "epaper.ClearScreen", "epaper")
	defer span.End()
	slog.DebugContext(ctx, "Clearing screen")
	device.ClearScreen()
}
