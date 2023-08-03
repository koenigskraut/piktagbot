package main

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	db "github.com/koenigskraut/piktagbot/database"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"os/signal"
)

var (
	botToken    = os.Getenv("BOT_TOKEN")
	sessionFile = os.Getenv("SESSION_FILE")
	appPort     = os.Getenv("APP_PORT")
	domain      = os.Getenv("DOMAIN")
)

func run(ctx context.Context) error {
	logger, _ := zap.NewDevelopment(zap.IncreaseLevel(zapcore.InfoLevel))
	defer func() { _ = logger.Sync() }()

	dispatcher := tg.NewUpdateDispatcher()
	fileSessionPath := sessionFile

	return telegram.BotFromEnvironment(
		ctx,
		telegram.Options{
			SessionStorage: &telegram.FileSessionStorage{Path: fileSessionPath},
			Logger:         logger,
			UpdateHandler:  dispatcher,
		}, func(ctx context.Context, client *telegram.Client) error {
			myClient := tg.NewClient(client)
			dispatcher.OnNewMessage(handleMessages(myClient))
			dispatcher.OnBotInlineQuery(handleInline(myClient))
			dispatcher.OnBotCallbackQuery(handleCallback(myClient))
			return nil
		},
		telegram.RunUntilCanceled)
}

func main() {
	db.InitializeDB()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
