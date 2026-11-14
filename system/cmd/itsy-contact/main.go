package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"itsy/internal/appruntime"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if appruntime.RunContact(ctx) != nil {
		os.Exit(1)
	}
}
