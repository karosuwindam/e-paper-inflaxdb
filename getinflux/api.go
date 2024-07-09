package getinflux

import "time"

type api struct {
	influx influxDBPass
}

func Init(url, db, table string) (*api, error) {
	influx := influxDBPass{
		url, db, table,
	}
	return &api{
		influx: influx,
	}, nil
}

func (a *api) InfluxdbDay(dayago int, datatype string) DataFormat {
	return getInfluxdbData(a.influx, dayago, datatype)
}

func (a *api) InfluxdbBack(backdate time.Duration, datatype string) DataFormat {
	return getInfluxdbData(a.influx, backdate, datatype)
}
