package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/otaviokr/go-epaper-lib"
	"golang.org/x/sync/errgroup"
)

var (
	M2in7bw = epaper.Model{Width: 176, Height: 264, StartTransmission: 0x13}
)

type infxseries struct {
	Name    string        `json:"name"`
	Columns []string      `json:"columns"`
	Values  []interface{} `json:"values"`
}

type infxresults struct {
	Statement_id int          `json:"statement_id"`
	Series       []infxseries `json:"series"`
}

type infxdata struct {
	Results []infxresults `json:"results"`
}

type dataformat struct {
	max float64
	avg float64
	min float64
}

func influxdb6h(datatype string) dataformat {
	// urldata := "http://192.168.0.6:8086/query?db=senser"
	urldata := "http://192.168.0.6:8086/query?db=senser"
	// datatype := "co2"
	// datatype := "tmp"
	body := "SELECT * FROM senser_data WHERE time >= '" + time.Now().Add(-time.Hour*6).UTC().Format("2006-01-02T15:04:05Z") + "' AND type='" + datatype + "'"
	urldata += "&q=" + url.QueryEscape(body)
	// values := url.Values{}
	// // values.Set("db", "senser")
	// values.Set("q", body)

	req, err := http.NewRequest("GET",
		urldata,
		// strings.NewReader(values.Encode()),
		nil,
	)
	if err != nil {

	}

	client := new(http.Client)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	var jsondata infxdata
	json.Unmarshal(byteArray, &jsondata)
	// fmt.Println(string(byteArray))
	tmpdata := jsondata.Results[0].Series[0]
	vid := 0
	for ; vid < len(tmpdata.Columns); vid++ {
		if tmpdata.Columns[vid] == "value" {
			break
		}
	}
	var tdata []float64
	for _, tmp := range tmpdata.Values {
		v := reflect.ValueOf(tmp).Index(vid).Interface()
		tdata = append(tdata, v.(float64))
	}
	ave := 0.0
	max := 0.0
	min := 0.0
	for i, num := range tdata {
		ave = (ave*float64(i) + num) / float64(i+1)
		if i == 0 {
			max = num
			min = num
		}
		if max < num {
			max = num
		}
		if min > num {
			min = num
		}
	}

	return dataformat{max: max, avg: math.Round(ave*100) / 100, min: min}

}
func influxdb10s(datatype string) dataformat {
	// urldata := "http://192.168.0.6:8086/query?db=senser"
	urldata := "http://192.168.0.6:8086/query?db=senser"
	// datatype := "co2"
	// datatype := "tmp"
	body := "SELECT * FROM senser_data WHERE time >= '" + time.Now().Add(-time.Minute*10).UTC().Format("2006-01-02T15:04:05Z") + "' AND type='" + datatype + "'"
	urldata += "&q=" + url.QueryEscape(body)
	// values := url.Values{}
	// // values.Set("db", "senser")
	// values.Set("q", body)

	req, err := http.NewRequest("GET",
		urldata,
		// strings.NewReader(values.Encode()),
		nil,
	)
	if err != nil {

	}

	client := new(http.Client)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)
	var jsondata infxdata
	json.Unmarshal(byteArray, &jsondata)
	// fmt.Println(string(byteArray))
	tmpdata := jsondata.Results[0].Series[0]
	vid := 0
	for ; vid < len(tmpdata.Columns); vid++ {
		if tmpdata.Columns[vid] == "value" {
			break
		}
	}
	var tdata []float64
	for _, tmp := range tmpdata.Values {
		v := reflect.ValueOf(tmp).Index(vid).Interface()
		tdata = append(tdata, v.(float64))
	}
	ave := 0.0
	max := 0.0
	min := 0.0
	for i, num := range tdata {
		ave = (ave*float64(i) + num) / float64(i+1)
		if i == 0 {
			max = num
			min = num
		}
		if max < num {
			max = num
		}
		if min > num {
			min = num
		}
	}

	return dataformat{max: max, avg: math.Round(ave*100) / 100, min: min}

}

func main() {

	epd, err := ESetup()

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	epd.Init()

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		for {
			co2data := influxdb10s("co2")
			tmpdata := influxdb10s("tmp")
			tmpdata6h := influxdb6h("tmp")
			humdata := influxdb10s("hum")
			humdata6h := influxdb6h("hum")
			output := []string{
				"気温",
				fmt.Sprintf(" 現時点:%v", tmpdata.avg),
				fmt.Sprintf(" 平均　:%v", tmpdata6h.avg),
				fmt.Sprintf(" 最大:%v", tmpdata6h.max),
				fmt.Sprintf(" 最小:%v", tmpdata6h.min),
				"",
				"湿度",
				fmt.Sprintf(" 現時点:%v", humdata.avg),
				fmt.Sprintf(" 平均　:%v", humdata6h.avg),
				"",
				"CO2",
				fmt.Sprintf(" 現時点:%v", co2data.avg),
			}
			epd.ClearScreen()
			textPut(epd, 0, 0, output, 20)
			time.Sleep(time.Minute * 5)
		}
	})
	<-ctx.Done()
	fmt.Println("shutdown")

	// textPut(epd, 0, 20, []string{"こんにちは", "こんにちは"}, 20)
}
