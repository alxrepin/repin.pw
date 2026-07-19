package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	app "repin/internal/bootstrap/http"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	application := app.New(ctx)

	go func() {
		<-ctx.Done()

		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := application.Stop(shutdown); err != nil {
			log.Println("graceful shutdown failed:", err)
		}
	}()

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
