package getprometheus

import (
	"context"
	"epaperifdb/config"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestGetprometheusJsonData(t *testing.T) {
	config.Init()
	// url := "http://prometheus.bookserver.home/api/v1/query_range?query=senser_data&start=2024-12-01T00:00:00.000Z&end=2024-12-01T00:15:00.000Z&step=15s"
	ctx := context.Background()
	ctx = contextWriteReadUrlData(ctx, 6*time.Hour)

	url := createPrometheusUrl(ctx, "senser_data")
	fmt.Println(url)

	ctx = contextWriteUrl(ctx, url)
	pdata := getprometheusJsonData(ctx)
	data := convertPrometheusData(ctx, pdata)
	fmt.Println(data)
	cvalue := data.convertTypeValue(ctx)
	fmt.Println(cvalue)

	ctx = contextWriteDataName(ctx, "co2")
	fmt.Println("co2", cvalue.convertData(ctx))

	ctx = contextWriteDataName(ctx, "tmp")
	fmt.Println("tmp", cvalue.convertData(ctx))

	ctx = contextWriteDataName(ctx, "tmp2")
	fmt.Println("tmp2", cvalue.convertData(ctx))
}

func TestGetprometheus2(t *testing.T) {
	os.Setenv("URL_PROM", "http://localhost:9090")
	config.Init()
	ctx := context.Background()
	ctx = contextWriteReadUrlData(ctx, 6*time.Hour)

	url := createPrometheusUrl(ctx, "senser_hum_value")
	fmt.Println(url)

	ctx = contextWriteUrl(ctx, url)
	pdata := getprometheusJsonData(ctx)
	data := convertPrometheusData(ctx, pdata)
	fmt.Println(data)
	cvalue := data.convertTypeValue(ctx)
	fmt.Println(cvalue)
	tmps := readVlue(ctx, "senser_tmp_value")
	cvalue = margeValue(cvalue, tmps)
	tmps = readVlue(ctx, "senser_press_value")
	cvalue = margeValue(cvalue, tmps)

	ctx = contextWriteDataName(ctx, "hum")
	fmt.Println("hum", cvalue.convertData(ctx))
	ctx = contextWriteDataName(ctx, "tmp")
	fmt.Println("tmp", cvalue.convertData(ctx))

}
