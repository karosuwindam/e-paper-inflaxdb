package epaperv2

import (
	"bytes"
	"context"
	"epaperifdb/config"
	"image"
	"log/slog"
)

type api struct {
	device Epd
}

func Init() (*api, error) {
	e := CreateEpd()
	if err := e.Open(); err != nil {
		return &api{}, err
	}
	defer e.Close()
	e.Init()

	if config.OutURL.InitClear {
		e.Clear()
	}
	return &api{device: e}, nil
}

func (a *api) TextPut(ctx context.Context, x, y int, texts []string, size float64) error {
	return a.device.testPut(ctx, x, y, texts, size)
}

func (a *api) ClearScreen(ctx context.Context) {
	ctx, span := config.TracerS(ctx, "epaper.ClearScreen", "epaper")
	defer span.End()
	slog.DebugContext(ctx, "Clearing screen")
	if err := a.device.Open(); err != nil {
		slog.ErrorContext(ctx, "spi deice open error", "error", err)
		return
	}
	defer a.device.Close()
	a.device.Clear()
	// a.device.CrearDisplayData()
}

func (a *api) CrearDisplayData(ctx context.Context) {
	ctx, span := config.TracerS(ctx, "epaper.ClearDisplayData", "epaper")
	defer span.End()
	slog.DebugContext(ctx, "Crearing Display Data")
	a.device.CrearDisplayData()
}

func (e *Epd) testPut(ctx context.Context, x, y int, texts []string, size float64) error {
	ctx, span := config.TracerS(ctx, "epaper.testPut", "epaper")
	defer span.End()
	slog.DebugContext(ctx, "testPut", "x", x, "y", y, "text", texts, "size", size)

	ctx = contextWriteWriteData(ctx, texts, size)
	bufferReader := bytes.NewReader(writedata(ctx))
	image, _, err := image.Decode(bufferReader)
	if err != nil {
		// FIXME Better error handling.
		return err
	}
	err = e.Open()
	if err != nil {
		return err
	}
	defer e.Close()
	e.AddLayer(image, x, y, true)
	e.PrintDisplay(config.OutURL.Rotate180)
	return nil
}
