package commands

import (
	"context"
	"errors"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

type PrePostDefHandler func(context.Context, tg.Entities, *tg.UpdateNewMessage, *HelperCapture) error
type CommandHandler func(context.Context, tg.Entities, *tg.UpdateNewMessage, *HelperCapture, string) error

type preDefaultHandlers struct {
	pre  PrePostDefHandler
	post PrePostDefHandler
	def  PrePostDefHandler
}

// HelperCapture contains API client and various helpers that can be provided to CommandDispatcher by With*()
// method. Not intended to be used directly, only for field access.
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
	UserCapture any
}

// CommandDispatcher is a bot message handling dispatcher that also contains a capture with helpers and arbitrary
// user data (HelperCapture.UserCapture)
type CommandDispatcher struct {
	handlers       map[string]CommandHandler
	prePostDefault *preDefaultHandlers
	capture        *HelperCapture
}

// ErrNoAction is a signal that message was not consumed by pre handler, returning it is the default and intended
// behaviour, use it in your custom pre handler.
var ErrNoAction = errors.New("no action")

// ErrDoNotProcess signalize that message was not processed, so Post won't be executed. useful if there is some cleanup in Post handler
var ErrDoNotProcess = errors.New("message is skipped")

// NewCommandDispatcher creates new bot dispatcher and attaches itself to the provided tg.UpdateDispatcher.
// If it is not the intended behaviour, you can provide nil as an argument and get message handler with
// CommandDispatcher.NewMessageHandler.
func NewCommandDispatcher(handler *tg.UpdateDispatcher) CommandDispatcher {
	c := CommandDispatcher{
		handlers: map[string]CommandHandler{},
		prePostDefault: &preDefaultHandlers{
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

// Pre sets-up a handler, that will be executed before any command is parsed, default behaviour: return ErrNoAction.
func (u CommandDispatcher) Pre(handler PrePostDefHandler) CommandDispatcher {
	u.prePostDefault.pre = handler
	return u
}

// Post sets-up a handler, that will be executed after all handlers
func (u CommandDispatcher) Post(handler PrePostDefHandler) CommandDispatcher {
	u.prePostDefault.post = handler
	return u
}

// Default sets-up a handler, that will be executed if no commands are recognized.
func (u CommandDispatcher) Default(handler PrePostDefHandler) CommandDispatcher {
	u.prePostDefault.def = handler
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

func (u CommandDispatcher) dispatch(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) (err error) {
	if update == nil {
		return nil
	}
	msg, ok := update.Message.(*tg.Message)
	if !ok {
		return nil
	}

	preErr := u.prePostDefault.pre(ctx, e, update, u.capture)
	switch preErr {
	case ErrDoNotProcess: // not a message of interest, immediate return
		return nil
	case ErrNoAction: // normal behaviour, defer post, continue
		defer func() {
			postErr := u.prePostDefault.post(ctx, e, update, u.capture)
			err = errors.Join(err, postErr)
		}()
	case nil: // message consumed by pre, execute post & return
		return u.prePostDefault.post(ctx, e, update, u.capture)
	default: // other error, execute post anyway, join errors and return
		postErr := u.prePostDefault.post(ctx, e, update, u.capture)
		return errors.Join(preErr, postErr)
	}

	cmd, ok := readFirstCommand(msg)
	if !ok {
		return u.prePostDefault.def(ctx, e, update, u.capture)
	}
	handler, ok := u.handlers[cmd.command]
	if !ok {
		err = u.prePostDefault.def(ctx, e, update, u.capture)
	} else {
		err = handler(ctx, e, update, u.capture, cmd.clear)
	}
	return
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
