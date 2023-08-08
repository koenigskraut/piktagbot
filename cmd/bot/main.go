package main

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/commands"
	db "github.com/koenigskraut/piktagbot/database"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"os/signal"
)

var (
	botToken    = os.Getenv("BOT_TOKEN")
	_           = botToken // compiler warning fix
	sessionFile = os.Getenv("SESSION_FILE")
	appDomain   = os.Getenv("DOMAIN")
	appPort     = os.Getenv("APP_PORT")
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

			cmdMap := map[string]commands.CommandHandler{
				"start":  commands.Start,
				"help":   commands.Help,
				"cancel": commands.Cancel,
				"tag":    commands.Tag,
				"remove": commands.Remove,
				"global": commands.Global,
			}

			cmdDispatcher := commands.NewCommandDispatcher(&dispatcher).
				WithClient(myClient).
				WithSender(message.NewSender(myClient)).
				WithUploader(uploader.NewUploader(myClient)).
				WithDownloader(downloader.NewDownloader()).
				WithCommands(cmdMap)
			cmdDispatcher.Pre(handlePre())

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
