package getinflux

import (
	"context"
	"time"
)

type api struct {
	influx *influxDBPass
}

func Init(url, db, table string) (*api, error) {
	influx := &influxDBPass{
		url, db, table,
	}
	return &api{
		influx: influx,
	}, nil
}

func (a *api) InfluxdbDay(ctx context.Context, dayago int, datatype string) DataFormat {
	return a.influx.getInfluxdbData(ctx, dayago, datatype)
}

func (a *api) InfluxdbBack(ctx context.Context, backdate time.Duration, datatype string) DataFormat {
	return a.influx.getInfluxdbData(ctx, backdate, datatype)
}
