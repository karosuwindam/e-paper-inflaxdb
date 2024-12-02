package main

import (
	"context"
	"epaperifdb/controller"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	if err := controller.Init(); err != nil {
		panic(err)
	}
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
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
