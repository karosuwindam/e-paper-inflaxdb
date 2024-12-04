package controller

import (
	"context"
	"epaperifdb/config"
	"epaperifdb/controller/commondata"
	"epaperifdb/controller/epaper"
	getinflux "epaperifdb/controller/getInflux"
	"epaperifdb/controller/getprometheus"
	"errors"
	"fmt"
	"log/slog"
	"sync"
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

	var co2data, tmpdata, tmpdata6h, tmpdate1d, humdata, humdata6h commondata.DataFormat

	epdApi, err := epaper.Init()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to initialize e-paper device", "error", err.Error())
		return err
	}
	//prometheusのデータ読み取り
	var wg sync.WaitGroup
	wg.Add(3)
	var data1day chan map[string]commondata.DataFormat = make(chan map[string]commondata.DataFormat, 1)
	go func(ctx context.Context) {
		defer wg.Done()
		ctx, span1 := config.TracerS(ctx, "PrometheusRead1day", "Promethesu Read 1day")
		defer span1.End()
		tmpdata1day, terr := getprometheus.GetPrometheusDays(ctx, 1)
		if terr != nil {
			slog.ErrorContext(ctx, "getpromethuesDay", "error", terr)
			return
		}
		tmp := map[string]commondata.DataFormat{
			"tmp": tmpdata1day.ConvertTmp(ctx),
		}
		data1day <- tmp
	}(ctx)
	var data6hour chan map[string]commondata.DataFormat = make(chan map[string]commondata.DataFormat, 1)
	go func(ctx context.Context) {
		defer wg.Done()
		ctx, span1 := config.TracerS(ctx, "PrometheusRead1day", "Promethesu Read 6 Hour")
		defer span1.End()
		tmpdata6hour, terr := getprometheus.GetPrometheusBack(ctx, 6*time.Hour)
		if terr != nil {
			slog.ErrorContext(ctx, "getpromethuesDay", "error", terr)
			return
		}
		tmp := map[string]commondata.DataFormat{
			"tmp": tmpdata6hour.ConvertTmp(ctx),
			"hum": tmpdata6hour.ConvertHum(ctx),
		}
		data6hour <- tmp
	}(ctx)
	var data10min chan map[string]commondata.DataFormat = make(chan map[string]commondata.DataFormat, 1)
	go func(ctx context.Context) {
		defer wg.Done()
		ctx, span1 := config.TracerS(ctx, "PrometheusRead1day", "Promethesu Read 10 Minute")
		defer span1.End()
		tmpdata10min, terr := getprometheus.GetPrometheusBack(ctx, 10*time.Minute)
		if terr != nil {
			slog.ErrorContext(ctx, "getpromethuesDay", "error", terr)
			return
		}
		tmp := map[string]commondata.DataFormat{
			"tmp": tmpdata10min.ConvertTmp(ctx),
			"co2": tmpdata10min.ConvertCO2(ctx),
			"hum": tmpdata10min.ConvertHum(ctx),
		}
		data10min <- tmp
	}(ctx)

	inluxApi, err := getinflux.Init(config.OutURL.InfluxDBUrl, DB, TABLE)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to initialize InfluxDB", "error", err.Error())
		return err
	}
	ctxInlux, cancel := context.WithCancel(ctx)
	co2data = inluxApi.InfluxdbBack(ctxInlux, time.Minute*10, "co2")
	tmpdata = inluxApi.InfluxdbBack(ctxInlux, time.Minute*10, "tmp")
	if co2data.Avg == 0 && tmpdata.Avg == 0 {
		slog.WarnContext(ctx, "InfluxData Not read Data to cannceled",
			"co2data", co2data,
			"tmpdata", tmpdata,
		)
		cancel()
	}
	tmpdata6h = inluxApi.InfluxdbBack(ctxInlux, time.Hour*6, "tmp")
	tmpdate1d = inluxApi.InfluxdbDay(ctxInlux, 1, "tmp")
	humdata = inluxApi.InfluxdbBack(ctxInlux, time.Minute*10, "hum")
	humdata6h = inluxApi.InfluxdbBack(ctxInlux, time.Hour*6, "hum")
	if tmpdate1d.Avg == 0 && tmpdata.Max == 0 {
		slog.WarnContext(ctx, "InfluxData Not read Data",
			"co2data", co2data,
			"tmpdata", tmpdata,
			"tmpdata6h", tmpdata6h,
			"tmpdate1d", tmpdate1d,
			"humdata", humdata,
			"humdata6h", humdata6h,
		)
		wg.Wait()
		if len(data10min) == 0 || len(data1day) == 0 || len(data6hour) == 0 {
			return errors.New("prometheus Not read Data")
		}
		tmp := <-data10min
		slog.DebugContext(ctx, "Read Senser Data for 10 min", "tmp", tmp)
		co2data = tmp["co2"]
		tmpdata = tmp["tmp"]
		humdata = tmp["hum"]
		tmp = <-data6hour
		slog.DebugContext(ctx, "Read Senser Data for 6 hour", "tmp", tmp)
		tmpdata6h = tmp["tmp"]
		humdata6h = tmp["hum"]
		tmp = <-data1day
		slog.DebugContext(ctx, "Read Senser Data for 1 day", "tmp", tmp)
		tmpdate1d = tmp["tmp"]
	}
	close(data10min)
	close(data1day)
	close(data6hour)

	slog.DebugContext(ctx, "Read Senser Data",
		"co2data", co2data,
		"tmpdata", tmpdata,
		"tmpdata6h", tmpdata6h,
		"tmpdate1d", tmpdate1d,
		"humdata", humdata,
		"humdata6h", humdata6h,
	)

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
