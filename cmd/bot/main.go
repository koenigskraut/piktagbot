package main

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	cmd "github.com/koenigskraut/piktagbot/commands"
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

			cmdMap := map[string]cmd.CommandHandler{
				"start":  cmd.Start,
				"help":   cmd.Help,
				"cancel": cmd.Cancel,
				"tag":    cmd.Tag,
				"remove": cmd.Remove,
				"global": cmd.Global,
				"check":  cmd.Check,
			}

			cmdDispatcher := cmd.NewCommandDispatcher(&dispatcher).
				WithClient(myClient).
				WithCommands(cmdMap)
			cmdDispatcher.Pre(handlePre())
			cmdDispatcher.Post(handlePost)

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
