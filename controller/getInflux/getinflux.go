package getinflux

import (
	"context"
	"encoding/json"
	"epaperifdb/config"
	"errors"
	"fmt"
	"io/ioutil"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type influxDBPass struct {
	url   string
	db    string
	table string
}

type strReadUrlData struct {
	timeAgo  interface{}
	dataType string
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
	CKDATAP       = 80
	ClientTImeout = 5
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

func jsonDataToDataformat(jsonData infxData) (DataFormat, error) {
	slog.DebugContext(context.Background(), "jsonDataToDataformat", "jsonData", jsonData)
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

func getInfluxJsonData(ctx context.Context) (infxData, error) {
	ctx, span := config.TracerS(ctx, "getInfluxJsonData", "get influx json data")
	defer span.End()
	var jsondata infxData

	ctx, cancel := context.WithTimeout(ctx, time.Second*ClientTImeout)
	defer cancel()

	urlData, ok := contextReadUrl(ctx)
	if !ok {
		slog.ErrorContext(ctx, "contextReadUrl error")
		return jsondata, fmt.Errorf("contextReadUrl error")
	}
	slog.DebugContext(ctx, "getInfluxJsonData", "urlData", urlData)

	req, err := http.NewRequestWithContext(ctx, "GET", urlData, nil)
	if err != nil {
		slog.ErrorContext(ctx, "http.NewRequestWithContext error", "error", err)
		return jsondata, err
	}
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "http.Do error", "error", err)
		return jsondata, err
	}
	defer resp.Body.Close()
	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		slog.ErrorContext(ctx, "ioutil.ReadAll error", "error", err)
		return jsondata, err
	}
	if err := json.Unmarshal(byteArray, &jsondata); err != nil {
		slog.ErrorContext(ctx, "json.Unmarshal error", "error", err)
		return jsondata, err
	}
	return jsondata, nil
}

func (influxDB *influxDBPass) getInfluxdbData(ctx context.Context, timeAgo interface{}, dataType string) DataFormat {
	ctx, span := config.TracerS(ctx, "getInfluxdbData", "get influxdb data")
	defer span.End()
	slog.DebugContext(ctx, "getInfluxdbData", "timeAgo", timeAgo, "dataType", dataType)

	ctx = contextWriteReadUrlData(ctx, timeAgo, dataType)
	url := influxDB.createReadUrlData(ctx)

	ctx = contextWriteUrl(ctx, url)
	jsonData, err := getInfluxJsonData(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "getInfluxJsonData error", "error", err)
		return DataFormat{}
	}
	data, err := jsonDataToDataformat(jsonData)
	if err != nil {
		slog.ErrorContext(ctx, "jsonDataToDataformat error", "error", err)
		return DataFormat{}
	}
	return data
}

func (influxDB *influxDBPass) createReadUrlData(ctx context.Context) string {
	ctx, span := config.TracerS(ctx, "createReadUrlData", "read url data")
	defer span.End()
	timedata, ok := contextReadReadUrlData(ctx)
	if !ok {
		slog.ErrorContext(ctx, "contextReadReadUrlData error")
		return ""
	}
	slog.DebugContext(ctx, "createReadUrlData", "timedata", timedata)

	urlData := fmt.Sprintf("%v/query?db=%v", influxDB.url, influxDB.db)
	bodyData := ""
	switch timedata.timeAgo.(type) {
	case int:
		t := time.Now().Add(-time.Hour * 24 * time.Duration(timedata.timeAgo.(int)))
		aTime := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)

		bodyData = fmt.Sprintf(
			"SELECT * FROM %v WHERE time <='%v' AND time >= '%v' AND type = '%v'",
			influxDB.table,
			aTime.UTC().Format("2006-01-02T15:04:05Z"),
			aTime.Add(-time.Hour*24).UTC().Format("2006-01-02T15:04:05Z"),
			timedata.dataType,
		)
	case time.Duration:
		t := timedata.timeAgo.(time.Duration)
		bodyData = fmt.Sprintf(
			"SELECT * FROM %v WHERE time >='%v' AND type = '%v'",
			influxDB.table,
			time.Now().Add(-t).UTC().Format("2006-01-02T15:04:05Z"),
			timedata.dataType,
		)
	default:
		slog.ErrorContext(ctx, "timeAgo type error")
		return ""
	}
	slog.DebugContext(ctx, "createReadUrlData",
		"url", fmt.Sprintf("%v&q=%v", urlData, url.QueryEscape(bodyData)),
		"bodyData", bodyData,
		"urlData", urlData,
	)
	return fmt.Sprintf("%v&q=%v", urlData, url.QueryEscape(bodyData))
}
