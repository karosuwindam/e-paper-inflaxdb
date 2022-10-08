package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	IP    = "192.168.0.6"
	Port  = "8086"
	DB    = "senser"
	TABLE = "senser_data"
)

func networkcheck() error {

	urldata := "http://" + IP + ":" + Port
	if req, err := http.NewRequest("GET",
		urldata,
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

func main() {
	epd, err := ESetup()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	for i := 0; i < 3; i++ {
		if i == 0 {
			time.Sleep(time.Millisecond * 500)
		} else {
			time.Sleep(time.Minute)
		}
		if err := networkcheck(); err == nil {
			break
		} else {
			log.Println(err)
			log.Println("err count:", i)
		}
		if i == 2 {
			fmt.Println("network err", IP, Port)
			return
		}
	}

	epd.Init()

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		for {
			co2data := influxdbBack(time.Minute*10, "co2")
			tmpdata := influxdbBack(time.Minute*10, "tmp")
			tmpdata6h := influxdbBack(time.Hour*6, "tmp")
			tmpdate1d := influxdbday(1, "tmp")
			humdata := influxdbBack(time.Minute*10, "hum")
			humdata6h := influxdbBack(time.Hour*6, "hum")
			if tmpdate1d.avg == 0 && tmpdata.max == 0 {
				time.Sleep(time.Microsecond * 500)
				continue
			}
			output := []string{
				time.Now().Format("15:04:05"),
				"-気温(昨日)",
				fmt.Sprintf(" 現時点:%.1f", tmpdata.avg),
				fmt.Sprintf(" 平均6h:%.1f(%.1f)", tmpdata6h.avg, tmpdate1d.avg),
				fmt.Sprintf(" 最大6h:%.1f(%.1f)", tmpdata6h.max, tmpdate1d.max),
				fmt.Sprintf(" 最小6h:%.1f(%.1f)", tmpdata6h.min, tmpdate1d.min),
				"-湿度",
				fmt.Sprintf(" 現時点:%.1f", humdata.avg),
				fmt.Sprintf(" 平均6h:%.1f", humdata6h.avg),
				"-CO2",
				fmt.Sprintf(" 現時点:%.1f", co2data.avg),
			}
			epd.ClearScreen()
			textPut(epd, 0, 0, output, 20)
			time.Sleep(time.Minute * 5)
		}
	})
	<-ctx.Done()
	fmt.Println("shutdown")

}
