package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"flea/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(cli.RunExecutable(ctx, os.Args[0], os.Args[1:], os.Stdout, os.Stderr))
}
