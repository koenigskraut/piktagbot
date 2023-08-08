package commands

import (
	"context"
	"errors"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

type CommandHandler func(context.Context, tg.Entities, *tg.UpdateNewMessage, *HelperCapture) error

type preDefaultHandlers struct {
	pre CommandHandler
	def CommandHandler
}

// HelperCapture contains API client and various helpers that can be provided to CommandDispatcher by With*()
// method.
//
// Example (updatesDispatcher is tg.UpdateDispatcher, client is *telegram.Client):
//
//	api := tg.NewClient(client)
//	cmdDispatcher := commands.NewCommandDispatcher(&updatesDispatcher).
//		WithClient(api).
//		WithSender(message.NewSender(api)).
//		WithUploader(uploader.NewUploader(api)).
//		WithDownloader(downloader.NewDownloader())
//
//	startCmd := func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage, c *commands.HelperCapture) error {
//		_, err := c.Sender.Answer(e, u).Text(ctx, "Hello there!")
//		return err
//	}
//	cmdDispatched.OnNewCommand("start", startCmd)
type HelperCapture struct {
	Client      *tg.Client
	Sender      *message.Sender
	Uploader    *uploader.Uploader
	Downloader  *downloader.Downloader
	Clear       string
	UserCapture any
}

// CommandDispatcher f
type CommandDispatcher struct {
	handlers   map[string]CommandHandler
	preDefault *preDefaultHandlers
	capture    *HelperCapture
}

// ErrNoAction is a signal that message was not consumed by pre handler, returning it is the default and intended
// behaviour, use it in your custom pre handler.
var ErrNoAction = errors.New("no action")

// NewCommandDispatcher creates new bot dispatcher and attaches itself to the provided tg.UpdateDispatcher.
// If it is not the intended behaviour, you can provide nil as an argument and get message handler with
// CommandDispatcher.NewMessageHandler.
func NewCommandDispatcher(handler *tg.UpdateDispatcher) CommandDispatcher {
	c := CommandDispatcher{
		handlers: map[string]CommandHandler{},
		preDefault: &preDefaultHandlers{
			pre: func(context.Context, tg.Entities, *tg.UpdateNewMessage, *HelperCapture) error { return ErrNoAction },
			def: func(context.Context, tg.Entities, *tg.UpdateNewMessage, *HelperCapture) error { return nil },
		},
		capture: &HelperCapture{},
	}
	handler.OnNewMessage(c.NewMessageHandler)
	return c
}

func (u CommandDispatcher) WithCommands(commands map[string]CommandHandler) CommandDispatcher {
	for k, v := range commands {
		u.handlers[k] = v
	}
	return u
}

func (u CommandDispatcher) WithClient(client *tg.Client) CommandDispatcher {
	u.capture.Client = client
	return u
}

func (u CommandDispatcher) WithSender(sender *message.Sender) CommandDispatcher {
	u.capture.Sender = sender
	return u
}

func (u CommandDispatcher) WithUploader(uploader *uploader.Uploader) CommandDispatcher {
	u.capture.Uploader = uploader
	u.capture.Sender = u.capture.Sender.WithUploader(uploader)
	return u
}

func (u CommandDispatcher) WithDownloader(downloader *downloader.Downloader) CommandDispatcher {
	u.capture.Downloader = downloader
	return u
}

// Default sets-up a handler, that will be executed if no commands are recognized.
func (u CommandDispatcher) Default(handler CommandHandler) CommandDispatcher {
	u.preDefault.def = handler
	return u
}

// Pre sets-up a handler, that will be executed before any command is parsed, default behaviour: return ErrNoAction.
func (u CommandDispatcher) Pre(handler CommandHandler) CommandDispatcher {
	u.preDefault.pre = handler
	return u
}

// OnNewCommand sets-up a handler for the command 'cmd'. No '/' is needed, only command name.
func (u CommandDispatcher) OnNewCommand(cmd string, handler CommandHandler) {
	u.handlers[cmd] = handler
}

// NewMessageHandler returns a handler for tg.UpdateDispatcher.OnNewMessage() method
func (u CommandDispatcher) NewMessageHandler(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
	return u.dispatch(ctx, e, update)
}

func (u CommandDispatcher) dispatch(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
	if update == nil {
		return nil
	}
	msg, ok := update.Message.(*tg.Message)
	if !ok {
		return nil
	}
	// either read error occurred or update was consumed (nil), return in both cases
	if err := u.preDefault.pre(ctx, e, update, u.capture); !errors.Is(err, ErrNoAction) {
		return err
	}
	cmd, ok := readFirstCommand(msg)
	if !ok {
		return u.preDefault.def(ctx, e, update, u.capture)
	}
	handler, ok := u.handlers[cmd.command]
	if !ok {
		handler = u.preDefault.def
	}
	u.capture.Clear = cmd.clear
	return handler(ctx, e, update, u.capture)
}

type readCommand struct {
	command string
	clear   string
}

func readFirstCommand(m *tg.Message) (command *readCommand, flag bool) {
	v, ok := m.MapEntities()
	if !ok {
		return nil, false
	}
	c, ok := v.AsMessageEntityBotCommand().First()
	if !ok {
		return nil, false
	}
	command = &readCommand{
		command: m.Message[c.Offset+1:][:c.Length-1],
		clear:   m.Message[0:c.Offset] + m.Message[c.Offset:][c.Length:],
	}
	flag = true
	return
}
