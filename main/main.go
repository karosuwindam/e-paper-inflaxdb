package main

import (
	"context"
	"epaperifdb/config"
	"epaperifdb/controller"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	if err := config.Init(); err != nil {
		panic(err)
	}
	if err := controller.Init(); err != nil {
		panic(err)
	}
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	tshutdown, terr := config.TracerStart(config.TraData.GrpcURL, config.TraData.ServiceName, ctx)
	if terr != nil {
		defer tshutdown(context.Background())
	}
	defer cancel()
	go func(ctx context.Context) {
		if err := controller.Run(ctx); err != nil {
			panic(err)
		}
	}(ctx)
	<-sigs
	if err := controller.Stop(context.Background()); err != nil {
		panic(err)
	}

}
