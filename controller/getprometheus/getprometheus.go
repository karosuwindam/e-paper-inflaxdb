package getprometheus

import (
	"context"
	"encoding/json"
	"epaperifdb/config"
	"epaperifdb/controller/commondata"
	"io/ioutil"
	"log/slog"
	"net/http"
	"sort"
	"strconv"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type prometheusData struct {
	Status string               `json:"status"`
	Data   prometheusDataStruct `json:"data"`
}

type prometheusDataStruct struct {
	ResultType string             `json:"resultType"`
	Result     []prometheusResult `json:"result"`
}

type prometheusResult struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"`
}

type typesPromethValue struct {
	value []prometheusValue
}

type prometheusValue struct {
	Timestamp time.Time
	Value     string
}

type strReadUrlData struct {
	timeAgo  interface{}
	dataType string
}

type convertTypeValues map[string]typesPromethValue

const (
	BASE_URL = "http://prometheus.bookserver.home/"
)

func createPrometheusUrl(ctx context.Context, query string) string {
	ctx, span := config.TracerS(ctx, "createPrometheusUrl", "crate prometheus url")
	defer span.End()
	slog.DebugContext(ctx, "createPrometheusUrl", "query", query)

	backdate := convertTimeAgoData(ctx)
	if backdate == 0 {
		slog.WarnContext(ctx, "createPrometheusUrl backdata == 0")
		return ""
	}
	baseUrl := config.OutURL.PrometheusUrl
	query = "senser_data"
	baseUrl += "/api/v1/query_range?query=" + query
	timenow := time.Now().UTC()
	baseUrl += "&start=" + timenow.Add(-backdate).Format("2006-01-02T15:04:05.000Z")
	baseUrl += "&end=" + timenow.Format("2006-01-02T15:04:05.000Z")
	baseUrl += "&step=15s"
	return baseUrl
}

func convertTimeAgoData(ctx context.Context) time.Duration {
	ctx, span := config.TracerS(ctx, "convertTimeAgoData", "convert Time Ago Data")
	defer span.End()
	slog.DebugContext(ctx, "convertTimeAgoData")

	timeAgo, ok := contextReadReadUrlData(ctx)
	if !ok {
		slog.WarnContext(ctx, "context not input data for timeAgo")
		return 0
	}
	switch timeAgo.(type) {
	case int: //days
		day := timeAgo.(int)
		return time.Hour * 24 * time.Duration(day)
	case time.Duration:
		timed := timeAgo.(time.Duration)
		return timed
	}
	slog.WarnContext(ctx, "timeAgo data err Type", "timeAgo", timeAgo)
	return 0
}

func convertPrometheusData(ctx context.Context, data []byte) prometheusData {
	ctx, span := config.TracerS(ctx, "convertPrometheusData", "convert Prometheus Data")
	defer span.End()
	slog.DebugContext(ctx, "convertPrometheusData", "data", string(data))

	var prometheusDatas prometheusData
	err := json.Unmarshal(data, &prometheusDatas)
	if err != nil {
		slog.ErrorContext(ctx, "json.Unmarshal error", "error", err)
		return prometheusData{}
	}
	return prometheusDatas
}

func (v *prometheusData) convertTypeValue(ctx context.Context) convertTypeValues {
	ctx, span := config.TracerS(ctx, "convertTypeValue", "convert Type Value")
	defer span.End()
	slog.DebugContext(ctx, "convertTypeValue", "prometheusData.Status", v.Status)

	value := make(map[string]typesPromethValue)
	if len(v.Data.Result) == 0 {
		slog.WarnContext(ctx, "v.Data.Result not entry")
		return nil
	}
	for _, result := range v.Data.Result {
		var values []prometheusValue
		for _, value := range result.Values {
			unixtime := value[0].(float64)
			timestamp := time.Unix(int64(unixtime), 0)

			values = append(values, prometheusValue{
				Timestamp: timestamp,
				Value:     value[1].(string),
			})
		}
		if _, ok := value[result.Metric["type"]]; !ok {
			value[result.Metric["type"]] = typesPromethValue{value: values}
		} else {
			value[result.Metric["type"]] = typesPromethValue{value: append(value[result.Metric["type"]].value, values...)}
		}
		// timestampごとにソート
		sort.Slice(value[result.Metric["type"]].value, func(i, j int) bool {
			return value[result.Metric["type"]].value[i].Timestamp.Before(value[result.Metric["type"]].value[j].Timestamp)
		})
	}
	return value
}

func (v convertTypeValues) convertData(ctx context.Context) commondata.DataFormat {
	ctx, span := config.TracerS(ctx, "convertData", "convert Data")
	defer span.End()
	dataname, ok := contextReadDataName(ctx)
	if !ok {
		slog.WarnContext(ctx, "context not input ctxdata")
		return commondata.DataFormat{}
	}
	slog.DebugContext(ctx, "convertData", "dataname", dataname)

	var data commondata.DataFormat
	var sum float64
	var count float64
	if _, ok := v[dataname]; !ok {
		return data
	}
	for _, value := range v[dataname].value {
		tmp, err := strconv.ParseFloat(value.Value, 64)
		if err != nil {
			slog.Error("strconv.ParseFloat error", "error", err)
		}
		if data.Max < tmp {
			if tmp <= 100 && dataname == "tmp" {
				data.Max = tmp
			}
		}
		if data.Min > tmp {
			data.Min = tmp
		} else if data.Min == 0 {
			data.Min = tmp
		}

		sum += tmp
		count++
	}
	data.Avg = sum / count
	return data
}

func getprometheusJsonData(ctx context.Context) []byte {
	ctx, span := config.TracerS(ctx, "getprometheusJsonData", "get prometheus JsonData")
	defer span.End()
	urldata, ok := contextReadUrl(ctx)
	if !ok {
		slog.WarnContext(ctx, "context not input urldata")
		return []byte{}
	}

	slog.DebugContext(ctx, "getprometheusJsonData", "urldata", urldata)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	slog.DebugContext(ctx, "GET "+urldata)
	req, err := http.NewRequestWithContext(ctx, "GET", urldata, nil)
	if err != nil {
		slog.ErrorContext(ctx, "http.NewRequestWithContext error", "error", err)
		return []byte{}
	}
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	resp, err := client.Do(req)
	if err != nil {
		slog.ErrorContext(ctx, "http.Do error", "error", err)
		return []byte{}
	}
	defer resp.Body.Close()
	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		slog.ErrorContext(ctx, "ioutil.ReadAll error", "error", err)
		return []byte{}
	}
	return byteArray
}
