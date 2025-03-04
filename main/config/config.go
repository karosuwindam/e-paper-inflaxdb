package config

import "github.com/caarlos0/env/v6"

type TracerData struct {
	GrpcURL        string `env:"TRACER_GRPC_URL" envDefault:"otel-grpc.bookserver.home:4317"`
	ServiceName    string `env:"TRACER_SERVICE_URL" envDefault:"e-paper"`
	TracerUse      bool   `env:"TRACER_ON" envDefault:"true"`
	ServiceVersion string `env:"TRACER_SERVICE_VERSION" envDefault:"0.0.4"`
}

type OutURLData struct {
	PrometheusUrl string `env:"URL_PROM" envDefault:"http://prometheus.bookserver.home"`
	// PrometheusUrl string `env:"URL_PROM" envDefault:"http://localhost:9090"`
	InfluxDBUrl string `env:"URL_INFLUX" envDefault:"http://192.168.0.6:8086"`
	ModuleType  bool   `env:"V2_FLAG" envDefault:"true"`
	Mirror      bool   `env:"V2_IMG_MIRROR" envDefault:"true"`
	InitClear   bool   `env:"INIT_CLEAR_FLAG" envDefault:"false"`
	Rotate180   bool   `env:"V2_ROTATE_180" envDefault:"true"`
}

var TraData TracerData
var OutURL OutURLData

func Init() error {
	TraData = TracerData{}

	if err := env.Parse(&TraData); err != nil {
		return err
	}
	if err := env.Parse(&OutURL); err != nil {
		return err
	}
	return nil
}
