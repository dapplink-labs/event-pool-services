package main

import (
	"context"
	"os"

	"github.com/ethereum/go-ethereum/log"

	"github.com/multimarket-labs/event-pod-services/common/opio"
)

func main() {
	log.SetDefault(log.NewLogger(log.NewTerminalHandlerWithLevel(os.Stdout, log.LevelInfo, true)))
	app := NewCli()
	ctx := opio.WithInterruptBlocker(context.Background())
	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Error("Application failed", "Err", err.Error())
		os.Exit(1)
	}
}
