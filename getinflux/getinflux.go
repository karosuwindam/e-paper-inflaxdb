package getinflux

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"sort"
	"time"
)

type influxDBPass struct {
	url   string
	db    string
	table string
}

type infxSeries struct {
	Name    string        `json:"name"`
	Columns []string      `json:"columns"`
	Values  []interface{} `json:"values"`
}

type infxresults struct {
	Statement_id int          `json:"statement_id"`
	Series       []infxSeries `json:"series"`
}

type infxData struct {
	Results []infxresults `json:"results"`
}

type DataFormat struct {
	Max float64
	Avg float64
	Min float64
}

type tmpNum struct {
	num float64
	c   int
}

const (
	CKDATAP = 80
)

func ckdata(data []float64) []float64 {
	tData := DataFormat{}
	for i, num := range data {
		tData.Avg = (tData.Avg*float64(i) + num) / float64(i+1)
		if i == 0 {
			tData.Max = num
			tData.Min = num
		}
		if tData.Max < num {
			tData.Max = num
		}
		if tData.Min > num {
			tData.Min = num
		}
	}
	var out []float64
	var cktmp []tmpNum
	for _, num := range data {
		tmp := math.Log10((num - tData.Min + 1) / (tData.Avg - tData.Min + 1))
		tmp = math.Ceil(math.Abs(tmp) * 10)
		tt := tmpNum{num: num, c: int(tmp)}
		cktmp = append(cktmp, tt)
	}
	sort.Slice(cktmp, func(i, j int) bool { return cktmp[i].c < cktmp[j].c })
	count := 0
	for i := 0; i < len(cktmp); i++ {
		if i < len(cktmp)*CKDATAP/100 {
			count = cktmp[i].c
		}
		if count == cktmp[i].c {
			out = append(out, cktmp[i].num)
		}

	}

	return out
}

func getInfluxdbData(influx influxDBPass, timeAgo interface{}, dataType string) DataFormat {
	// logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	urlData := createReadUrlData(influx, timeAgo, dataType)
	jsonData, err := getInfluxJsonData(urlData)
	if err != nil {
		logger.Error(err.Error(), "url", urlData)
		return DataFormat{}
	}
	d, err := jsonDataToDataformat(jsonData)
	if err != nil {
		logger.Error(err.Error())
		return DataFormat{}
	}
	return d
}

func jsonDataToDataformat(jsonData infxData) (DataFormat, error) {
	if len(jsonData.Results[0].Series) != 0 {
		tmpdata := jsonData.Results[0].Series[0]
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
		tdata = ckdata(tdata)
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
		return DataFormat{max, math.Round(ave*10) / 10, min}, nil
	}
	return DataFormat{}, errors.New("Input JsonData Not input")
}

func getInfluxJsonData(passUrl string) (infxData, error) {
	var jsondata infxData
	req, err := http.NewRequest("GET",
		passUrl,
		nil,
	)
	if err != nil {
		return jsondata, err
	}
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return jsondata, err
	}
	defer resp.Body.Close()
	byteArray, _ := ioutil.ReadAll(resp.Body)
	if err := json.Unmarshal(byteArray, &jsondata); err != nil {
		return jsondata, err
	}
	return jsondata, nil
}

func createReadUrlData(influx influxDBPass, timeAgo interface{}, dataType string) string {
	urlData := fmt.Sprintf("%v/query?db=%v", influx.url, influx.db)
	bodyData := ""
	switch timeAgo.(type) {
	case int:

		t := time.Now().Add(-time.Hour * 24 * time.Duration(timeAgo.(int)))
		aTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)

		bodyData = fmt.Sprintf(
			"SELECT * FROM %v WHERE time <='%v' AND time >= '%v' AND type = '%v'",
			influx.table,
			aTime.UTC().Format("2006-01-02T15:04:05Z"),
			aTime.Add(-time.Hour*24).UTC().Format("2006-01-02T15:04:05Z"),
			dataType,
		)
	case time.Duration:
		t := timeAgo.(time.Duration)
		bodyData = fmt.Sprintf(
			"SELECT * FROM %v WHERE time >='%v' AND type = '%v'",
			influx.table,
			time.Now().Add(-t).UTC().Format("2006-01-02T15:04:05Z"),
			dataType,
		)

	}
	return fmt.Sprintf("%v&q=%v", urlData, url.QueryEscape(bodyData))
}
