package main

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

func run(ctx context.Context) error {
	logger, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.InfoLevel))
	defer func() { _ = logger.Sync() }()

	dispatcher := tg.NewUpdateDispatcher()
	fileSessionPath := os.Getenv("SESSION_FILE")

	return telegram.BotFromEnvironment(
		ctx,
		telegram.Options{
			SessionStorage: &telegram.FileSessionStorage{Path: fileSessionPath},
			Logger:         logger,
			UpdateHandler:  dispatcher,
			NoUpdates:      true,
		}, func(ctx context.Context, client *telegram.Client) error {
			myClient := tg.NewClient(client)
			go downloadLoop(downloadChan, ctx, myClient)
			return nil
		},
		telegram.RunUntilCanceled)
}

func initializeTelegram(ctx context.Context) {
	err := run(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
