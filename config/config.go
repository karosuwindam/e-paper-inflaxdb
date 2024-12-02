package config

import "github.com/caarlos0/env/v6"

type TracerData struct {
	GrpcURL        string `env:"TRACER_GRPC_URL" envDefault:"otel-grpc.bookserver.home:4317"`
	ServiceName    string `env:"TRACER_SERVICE_URL" envDefault:"e-paper"`
	TracerUse      bool   `env:"TRACER_ON" envDefault:"true"`
	ServiceVersion string `env:"TRACER_SERVICE_VERSION" envDefault:"0.0.3"`
}

var TraData TracerData

func Init() error {
	TraData = TracerData{}

	if err := env.Parse(&TraData); err != nil {
		return err
	}
	// if err := logConfig(); err != nil {
	// 	return err
	// }
	return nil
}
