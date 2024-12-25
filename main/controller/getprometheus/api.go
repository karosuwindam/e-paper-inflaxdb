package getprometheus

import (
	"context"
	"epaperifdb/controller/commondata"
	"errors"
	"sync"
	"time"
)

type api struct {
	value convertTypeValues
}
type promeDatas map[string]typesPromethValue

func GetPrometheusDays(ctx context.Context, day int) (*api, error) {
	var wg sync.WaitGroup
	var chdata chan convertTypeValues = make(chan convertTypeValues, 2)
	wg.Add(2)
	go func(ctx context.Context, day int) {
		defer wg.Done()
		chdata <- readVlue(contextWriteReadUrlData(ctx, day), "senser_tmp_value")
	}(ctx, day)
	go func(ctx context.Context, day int) {
		defer wg.Done()
		chdata <- readVlue(contextWriteReadUrlData(ctx, day), "senser_hum_value")
	}(ctx, day)
	url := createPrometheusUrl(contextWriteReadUrlData(ctx, day), "senser_data")
	if url == "" {
		return nil, errors.New("input url data error")
	}

	pdata := getprometheusJsonData(contextWriteUrl(ctx, url))
	if len(pdata) == 0 {
		return nil, errors.New("not input data error")
	}
	data := convertPrometheusData(ctx, pdata)
	wg.Wait()
	v := data.convertTypeValue(ctx)
	for len(chdata) != 0 {
		tmp := <-chdata
		if tmp != nil {
			if v == nil {
				v = tmp
			} else {
				v = margeValue(v, tmp)
			}
		}
	}

	if v == nil {
		return nil, errors.New("not input read data error")

	}
	return &api{v}, nil

}

func GetPrometheusBack(ctx context.Context, backTime time.Duration) (*api, error) {
	var wg sync.WaitGroup
	var chdata chan convertTypeValues = make(chan convertTypeValues, 2)
	wg.Add(2)
	go func(ctx context.Context, backTime time.Duration) {
		defer wg.Done()
		chdata <- readVlue(contextWriteReadUrlData(ctx, backTime), "senser_tmp_value")
	}(ctx, backTime)
	go func(ctx context.Context, backTime time.Duration) {
		defer wg.Done()
		chdata <- readVlue(contextWriteReadUrlData(ctx, backTime), "senser_hum_value")
	}(ctx, backTime)
	url := createPrometheusUrl(contextWriteReadUrlData(ctx, backTime), "senser_data")
	if url == "" {
		return nil, errors.New("input url data error")
	}

	pdata := getprometheusJsonData(contextWriteUrl(ctx, url))
	if len(pdata) == 0 {
		return nil, errors.New("not input data error")
	}
	data := convertPrometheusData(ctx, pdata)
	wg.Wait()
	v := data.convertTypeValue(ctx)
	for len(chdata) != 0 {
		tmp := <-chdata
		if tmp != nil {
			if v == nil {
				v = tmp
			} else {
				v = margeValue(v, tmp)
			}
		}
	}
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
