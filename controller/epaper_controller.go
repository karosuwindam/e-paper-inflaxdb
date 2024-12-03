package controller

import (
	"context"
	"epaperifdb/config"
	"epaperifdb/controller/epaper"
	getinflux "epaperifdb/controller/getInflux"
	"fmt"
	"log/slog"
	"time"
)

const (
	IP       = "192.168.0.6"
	Port     = "8086"
	INFLUXDB = "http://192.168.0.6:8086"
	DB       = "senser"
	TABLE    = "senser_data"
)

func sensibleTemp(tmp, hum float64) float64 {
	return 37.0 - (37.0-tmp)/(0.68-0.0014*hum+1/(1.76+1.4)) - 0.29*tmp*(1-hum/100)
}

// e-paperの更新
func ePaperUpdate(ctx context.Context) error {
	ctx, span := config.TracerS(ctx, "ePaperUpdate", "e-paper update")
	defer span.End()
	slog.DebugContext(ctx, "ePaper Update Start")

	epdApi, err := epaper.Init()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to initialize e-paper device", "error", err.Error())
		return err
	}
	inluxApi, err := getinflux.Init(INFLUXDB, DB, TABLE)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to initialize InfluxDB", "error", err.Error())
		return err
	}

	co2data := inluxApi.InfluxdbBack(ctx, time.Minute*10, "co2")
	tmpdata := inluxApi.InfluxdbBack(ctx, time.Minute*10, "tmp")
	tmpdata6h := inluxApi.InfluxdbBack(ctx, time.Hour*6, "tmp")
	tmpdate1d := inluxApi.InfluxdbDay(ctx, 1, "tmp")
	humdata := inluxApi.InfluxdbBack(ctx, time.Minute*10, "hum")
	humdata6h := inluxApi.InfluxdbBack(ctx, time.Hour*6, "hum")
	if tmpdate1d.Avg == 0 && tmpdata.Max == 0 {
		slog.WarnContext(ctx, "Not read Data")
		return nil
	}

	output := []string{
		time.Now().Format("01/02 15:04:05"),
		"-気温(昨日)",
		fmt.Sprintf(" 現時点:%.1f", tmpdata.Avg),
		fmt.Sprintf(" 体感　:%.1f", sensibleTemp(tmpdata.Avg, humdata.Avg)),
		fmt.Sprintf(" 平均6h:%.1f(%.1f)", tmpdata6h.Avg, tmpdate1d.Avg),
		fmt.Sprintf(" 最大6h:%.1f(%.1f)", tmpdata6h.Max, tmpdate1d.Max),
		fmt.Sprintf(" 最小6h:%.1f(%.1f)", tmpdata6h.Min, tmpdate1d.Min),
		"-湿度",
		fmt.Sprintf(" 現時点:%.1f", humdata.Avg),
		fmt.Sprintf(" 平均6h:%.1f", humdata6h.Avg),
		"-CO2",
		fmt.Sprintf(" 現時点:%.1f", co2data.Avg),
	}

	epdApi.ClearScreen(ctx)
	if err := epdApi.TextPut(ctx, 0, 0, output, 20); err != nil {
		slog.ErrorContext(ctx, "Failed to put text", "error", err.Error())
		return err
	} else {
		slog.InfoContext(ctx, "ePaperUpdate")
	}
	slog.DebugContext(ctx, "ePaper Update End")

	return nil
}
