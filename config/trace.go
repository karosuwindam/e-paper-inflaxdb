package config

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	slogmulti "github.com/samber/slog-multi"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func initConn(url string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		url,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}
	return conn, err
}

func initResourc(ctx context.Context, servicName string) (*resource.Resource, error) {
	// res := resource.Default()
	res, err := resource.New(ctx,
		resource.WithFromEnv(), // Discover and provide attributes from OTEL_RESOURCE_ATTRIBUTES and OTEL_SERVICE_NAME environment variables.
		resource.WithProcessRuntimeName(),
		resource.WithContainer(),   // Discover and provide container information.
		resource.WithContainerID(), // Discover and provide container ID information.
		resource.WithHost(),        // Discover and provide host information.
	)
	if err != nil {
		return nil, err
	}

	return resource.Merge(res,
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(servicName),
			semconv.ServiceVersion(TraData.ServiceVersion),
		))
}

func initTracerProvider(ctx context.Context, res *resource.Resource, opt interface{}) (
	*sdktrace.TracerProvider, error) {
	var traceExporter *otlptrace.Exporter
	var err error
	if conn, ok := opt.(*grpc.ClientConn); ok {
		traceExporter, err = otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	} else if url, ok := opt.(string); ok {
		traceExporter, err = otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(url),
		)
	}
	if err != nil {
		return nil, err
	}
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter,
		sdktrace.WithBatchTimeout(time.Second),
	)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	return tracerProvider, nil
}

func initLoggerProvider(ctx context.Context, res *resource.Resource, opt interface{}) (
	*sdklog.LoggerProvider, error) {
	var logerExporter sdklog.Exporter
	var err error
	if conn, ok := opt.(*grpc.ClientConn); ok {
		logerExporter, err = otlploggrpc.New(ctx,
			otlploggrpc.WithGRPCConn(conn),
		)
	} else if url, ok := opt.(string); ok {
		logerExporter, err = otlploghttp.New(ctx,
			otlploghttp.WithEndpointURL(url),
		)
	}
	if err != nil {
		return nil, err
	}

	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logerExporter)),
	)
	return loggerProvider, nil
}

func initMetricProvider(ctx context.Context, res *resource.Resource, opt interface{}) (
	*sdkmetric.MeterProvider, error) {
	var metricProvider *sdkmetric.MeterProvider
	var metricExporter sdkmetric.Exporter
	var err error
	if conn, ok := opt.(*grpc.ClientConn); ok {
		metricExporter, err = otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	} else if url, ok := opt.(string); ok {
		metricExporter, err = otlpmetrichttp.New(ctx, otlpmetrichttp.WithEndpoint(url))
	}
	if err != nil {
		return nil, err
	}
	metricProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(metricExporter,
				sdkmetric.WithInterval(time.Second)),
		),
	)
	return metricProvider, nil
}

var tracer_ch func(context.Context) error

func TracerStart(urldata, serviceName string, ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	if !TraData.TracerUse {

		logConfig()
		return
	}
	slog.Info("Tracer Start", "url", urldata, "service", serviceName)
	conn, err := initConn(urldata)
	if err != nil {
		handleErr(err)
		return
	}
	res, err := initResourc(ctx, serviceName)
	if err != nil {
		handleErr(err)
		return
	}
	var tracerProvider *sdktrace.TracerProvider
	var meterProvider *sdkmetric.MeterProvider
	var loggerProvider *sdklog.LoggerProvider
	var errtmp error

	tracerProvider, errtmp = initTracerProvider(ctx, res, conn)
	if err != nil {
		handleErr(errtmp)
		return
	}
	loggerProvider, errtmp = initLoggerProvider(ctx, res, conn)
	if errtmp != nil {
		handleErr(errtmp)
		return
	}
	meterProvider, errtmp = initMetricProvider(ctx, res, conn)
	if errtmp != nil {
		handleErr(errtmp)
		return
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)
	logger := slog.New(
		slogmulti.Fanout(
			logHandler(logLevel),
			otelslog.NewHandler(
				serviceName,
				otelslog.WithSource(true),
			),
		),
	)
	slog.SetDefault(logger)

	return
}

func TracerStop(ctx context.Context) {
	if !TraData.TracerUse {
		return
	}
	if err := tracer_ch(ctx); err != nil {
		// log.Printf("error:", "failed to shutdown TracerProvider: %v", err)
		slog.ErrorContext(ctx, "failed to shutdown TracerProvider: %v", err)
	}
}

func TracerS(ctx context.Context, processName, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer(processName).Start(ctx, spanName, opts...)
}

func TraceHttpHandleFunc(mux *http.ServeMux, pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	// handler := otelhttp.WithRouteTag(pattern, http.HandlerFunc(handlerFunc))
	handler := otelhttp.NewHandler(http.HandlerFunc(handlerFunc), pattern,
		otelhttp.WithMessageEvents(otelhttp.ReadEvents, otelhttp.WriteEvents),
	)
	mux.Handle(pattern, handler)
}
