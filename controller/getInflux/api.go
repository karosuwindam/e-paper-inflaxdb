package getinflux

import (
	"context"
	"epaperifdb/config"
	"log/slog"
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
	ctx, span := config.TracerS(ctx, "Influxdb.Day", "influxdb")
	defer span.End()
	slog.DebugContext(ctx, "getInfluxdbData", "dayago", dayago, "datatype", datatype)
	return a.influx.getInfluxdbData(ctx, dayago, datatype)
}

func (a *api) InfluxdbBack(ctx context.Context, backdate time.Duration, datatype string) DataFormat {
	ctx, span := config.TracerS(ctx, "Influxdb.Day", "influxdb")
	defer span.End()
	slog.DebugContext(ctx, "getInfluxdbData", "backdate", backdate, "datatype", datatype)
	return a.influx.getInfluxdbData(ctx, backdate, datatype)
}
