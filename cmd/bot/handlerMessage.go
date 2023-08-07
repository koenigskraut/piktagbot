package main

import (
	"context"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/commands"
	"github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/flags"
	"sync"
)

func handlePre() func(context.Context, tg.Entities, *tg.UpdateNewMessage, *commands.HelperCapture) error {
	messageSemaphore := struct {
		sync.Mutex
		data map[int64]*sync.Mutex
	}{
		data: map[int64]*sync.Mutex{},
	}
	return func(ctx context.Context, entities tg.Entities, u *tg.UpdateNewMessage, c *commands.HelperCapture) error {
		m, ok := u.Message.(*tg.Message)
		// if there is an error or a message is outgoing/non-pm
		if !ok || m.Out || m.PeerID.TypeName() != "peerUser" {
			return nil
		}
		uID := m.PeerID.(*tg.PeerUser).UserID
		var currentLock *sync.Mutex

		// for every peer a new mutex is generated, mutexes are stored in a map
		// under the common map mutex
		messageSemaphore.Lock()
		if v, ok := messageSemaphore.data[uID]; !ok {
			currentLock = &sync.Mutex{}
			messageSemaphore.data[uID] = currentLock
		} else {
			currentLock = v
		}
		messageSemaphore.Unlock()

		// for this bot it is crucial to process messages synchronously
		currentLock.Lock()
		defer currentLock.Unlock()

		// get or create user record
		user := &database.User{UserID: uID}
		userIsNew, err := user.Get()
		if err != nil {
			return err
		}
		c.UserCapture = user

		// TODO get rid of strings, use enum-like constants, rework flag system
		// are we waiting for something from user?
		switch user.Flag {
		case "remove-tag":
			text, markup := flags.Remove(m, user)
			var err error
			if markup == nil {
				_, err = c.Sender.Answer(entities, u).Text(ctx, text)
			} else {
				_, err = c.Sender.Answer(entities, u).Markup(markup).Text(ctx, text)
			}
			return err
		case "add-sticker":
			text := flags.Add(m, user)
			_, err := c.Sender.Answer(entities, u).Text(ctx, text)
			return err
		default:
			break
		}

		// if not, check the message for commands:
		foundCommand, clearMessage := parseFirstCommand(m)
		if foundCommand != "" {
			answer := c.Sender.Answer(entities, u)
			switch foundCommand {
			case "start":
				return commands.Start(ctx, answer, userIsNew)
			case "help":
				return commands.Help(ctx, answer)
			case "global":
				return commands.Global(ctx, answer, user)
			case "cancel":
				return commands.Cancel(ctx, answer)
			case "tag":
				return commands.Tag(ctx, answer, m, c.Client, clearMessage)
			case "remove":
				return commands.Remove(ctx, answer, m, client)
			default:
				return nil
			}
		}
		return nil
	}
}

func parseFirstCommand(m *tg.Message) (command string, clearMessage string) {
	for _, e := range m.Entities {
		if e.TypeName() == "messageEntityBotCommand" {
			eA := e.(*tg.MessageEntityBotCommand)
			end := eA.Offset + eA.Length
			command = m.Message[eA.Offset+1 : end]
			clearMessage = m.Message[0:eA.Offset] + m.Message[end:]
			break
		}
	}
	return command, clearMessage
}
