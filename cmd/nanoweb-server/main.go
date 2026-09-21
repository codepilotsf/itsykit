package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"nanoweb/internal/host"
)

func main() {
	config := flag.String("config", "/etc/nanoweb/server.json", "server JSON configuration")
	check := flag.Bool("check", false, "check configuration and running server readiness")
	flag.Parse()
	logger := log.New(os.Stderr, "", 0)
	if flag.NArg() != 0 {
		logger.Print("unexpected arguments")
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var err error
	if *check {
		err = host.Check(ctx, *config)
	} else {
		err = host.Run(ctx, *config, logger)
	}
	if err != nil {
		logger.Print(err)
		os.Exit(1)
	}
}
