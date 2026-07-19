package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	app "repin/internal/bootstrap/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	application := app.New(ctx)
	defer application.Stop()

	if err := application.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
