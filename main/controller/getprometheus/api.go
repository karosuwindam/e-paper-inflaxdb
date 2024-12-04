package getprometheus

import (
	"context"
	"epaperifdb/controller/commondata"
	"errors"
	"time"
)

type api struct {
	value convertTypeValues
}
type promeDatas map[string]typesPromethValue

func GetPrometheusDays(ctx context.Context, day int) (*api, error) {
	url := createPrometheusUrl(contextWriteReadUrlData(ctx, day), "senser_data")
	if url == "" {
		return nil, errors.New("input url data error")
	}

	pdata := getprometheusJsonData(contextWriteUrl(ctx, url))
	if len(pdata) == 0 {
		return nil, errors.New("not input data error")
	}
	data := convertPrometheusData(ctx, pdata)
	v := data.convertTypeValue(ctx)
	if v == nil {
		return nil, errors.New("not input read data error")

	}
	return &api{v}, nil

}

func GetPrometheusBack(ctx context.Context, backTime time.Duration) (*api, error) {
	url := createPrometheusUrl(contextWriteReadUrlData(ctx, backTime), "senser_data")
	if url == "" {
		return nil, errors.New("input url data error")
	}

	pdata := getprometheusJsonData(contextWriteUrl(ctx, url))
	if len(pdata) == 0 {
		return nil, errors.New("not input data error")
	}
	data := convertPrometheusData(ctx, pdata)
	v := data.convertTypeValue(ctx)
	if v == nil {
		return nil, errors.New("not input read data error")

	}
	return &api{v}, nil

}

func (a *api) ConvertCO2(ctx context.Context) commondata.DataFormat {
	ctx = contextWriteDataName(ctx, "co2")
	return a.value.convertData(ctx)
}

func (a *api) ConvertTmp(ctx context.Context) commondata.DataFormat {
	ctx = contextWriteDataName(ctx, "tmp")
	return a.value.convertData(ctx)
}

func (a *api) ConvertHum(ctx context.Context) commondata.DataFormat {
	ctx = contextWriteDataName(ctx, "hum")
	return a.value.convertData(ctx)
}
