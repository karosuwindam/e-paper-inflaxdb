package main

import (
	"context"
	"epaperifdb/epaper"
	"epaperifdb/getinflux"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	IP       = "192.168.0.6"
	Port     = "8086"
	INFLUXDB = "http://192.168.0.6:8086"
	DB       = "senser"
	TABLE    = "senser_data"
)

func networkcheck() error {

	if req, err := http.NewRequest("GET",
		INFLUXDB,
		nil,
	); err != nil {
		return err
	} else {
		client := new(http.Client)
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	}
	return nil
}

func sensibleTemp(tmp, hum float64) float64 {
	return 37.0 - (37.0-tmp)/(0.68-0.0014*hum+1/(1.76+1.4)) - 0.29*tmp*(1-hum/100)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	var errch chan error = make(chan error, 1)
	var shutdown chan struct{} = make(chan struct{}, 1)
	go func() {
		i := 0
		for {
			if err := networkcheck(); err == nil {
				errch <- err
				break
			} else {
				errch <- err
			}
			i++
			if i > 3 {
				shutdown <- struct{}{}
				break
			}
			time.Sleep(time.Millisecond * 500)
		}
	}()
loop:
	for {
		select {
		case err := <-errch:
			if err == nil {
				logger.Info("InfluxDB Server Pass OK")
				break loop
			} else {
				logger.Warn(err.Error())
			}
		case <-shutdown:
			logger.Error("network err", "url", INFLUXDB)
			return
		}
	}
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go Run(ctx)
	<-sigs
	cancel()
	Stop()
	logger.Info("Process Shutdown")
}

func Stop() {
	select {
	case <-stopdone:
		break
	case <-time.After(time.Millisecond * 500):
		break
	}
	close(stopdone)
}

var stopdone chan struct{}

func Run(ctx context.Context) error {
	stopdone = make(chan struct{}, 1)
	if err := ePaperUpdate(); err != nil {
		return err
	}
loop:
	for {
		select {
		case <-ctx.Done():
			stopdone <- struct{}{}
			break loop
		case <-time.After(time.Minute * 5): //5分
			ePaperUpdate()
		}
	}
	return nil
}

func ePaperUpdate() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	epdApi, err := epaper.Init()
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	inluxApi, err := getinflux.Init(INFLUXDB, DB, TABLE)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	co2data := inluxApi.InfluxdbBack(time.Minute*10, "co2")
	tmpdata := inluxApi.InfluxdbBack(time.Minute*10, "tmp")
	tmpdata6h := inluxApi.InfluxdbBack(time.Hour*6, "tmp")
	tmpdate1d := inluxApi.InfluxdbDay(1, "tmp")
	humdata := inluxApi.InfluxdbBack(time.Minute*10, "hum")
	humdata6h := inluxApi.InfluxdbBack(time.Hour*6, "hum")
	if tmpdate1d.Avg == 0 && tmpdata.Max == 0 {
		logger.Warn("Not read Data")
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
	epdApi.ClearScreen()
	epdApi.TextPut(0, 0, output, 20)
	logger.Debug("e-Paper display update")
	return nil
}
