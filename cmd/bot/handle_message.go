package main

import (
	"context"
	"errors"
	"github.com/gotd/td/tg"
	cmd "github.com/koenigskraut/piktagbot/commands"
	db "github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/flags"
)

func handlePre() func(context.Context, tg.Entities, *tg.UpdateNewMessage, *cmd.HelperCapture) error {
	semaphore := cmd.NewMessageSemaphore()
	return func(ctx context.Context, entities tg.Entities, u *tg.UpdateNewMessage, c *cmd.HelperCapture) error {
		c.UserCapture = &semaphore
		m, ok := u.Message.(*tg.Message)
		// if there is an error or a message is outgoing/non-pm
		if !ok || m.Out || m.PeerID.TypeName() != "peerUser" {
			return cmd.ErrDoNotProcess
		}
		uID := m.PeerID.(*tg.PeerUser).UserID

		// for every peer a new mutex is generated, mutexes are stored in a map
		// under the common map mutex
		lockedUser := semaphore.GetCurrentLock(uID)

		// for this bot it is crucial to process messages synchronously for each user
		lockedUser.Lock()

		// get or create user record
		user := &db.User{UserID: uID}
		if err := user.Get(); err != nil {
			return err
		}
		lockedUser.DBUser = user

		answer := c.Sender.Answer(entities, u)
		// are we waiting for something from user?
		switch user.Flag {
		case flags.RemoveTag:
			return flags.Remove(ctx, m, user, answer)
		case flags.AddTag:
			return flags.AddOne(ctx, m, user, answer)
		case flags.AddTags:
			return flags.AddMany(ctx, m, user, answer)
		case flags.CheckTag:
			return flags.Check(ctx, m, user, answer)
		default:
			break
		}

		return cmd.ErrNoAction
	}
}

func handlePost(_ context.Context, _ tg.Entities, upd *tg.UpdateNewMessage, c *cmd.HelperCapture) error {
	ne, ok := upd.GetMessage().AsNotEmpty()
	if !ok {
		return errors.New("empty message")
	}
	semaphore := c.UserCapture.(*cmd.MessageSemaphore)
	var lockedUser *cmd.UserUnderLock
	switch v := ne.GetPeerID().(type) {
	case *tg.PeerUser:
		lockedUser = semaphore.GetCurrentLock(v.UserID)
	default:
		return errors.New("not a user chat")
	}
	lockedUser.Unlock()
	return nil
}
