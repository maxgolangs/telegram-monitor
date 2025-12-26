package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"getclient/internal/app"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	os.Exit(app.Run(ctx))
}



