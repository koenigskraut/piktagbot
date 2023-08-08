package commands

import (
	"context"
	_ "embed"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/tg"
)

//go:embed start_message.txt
var StartMessage string

func Start(ctx context.Context, e tg.Entities, upd *tg.UpdateNewMessage, c *HelperCapture, _ string) (err error) {
	answer := c.Sender.Answer(e, upd)
	_, user := c.UserCapture.(*MessageSemaphore).MessageUserFromUpdate(upd)
	if user.New {
		if _, err = answer.StyledText(ctx, html.String(nil, StartMessage)); err != nil {
			return err
		}
		user.New = false
		return user.Save()
	}
	_, err = answer.Text(ctx, "Вижу, Вы уже использовали бота, так что знаете что делать!"+
		"\nА если забыли, просто напишите /help")
	return
}
