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
	// TODO temporary regression, no new user check, add it to the DB table
	//if isNew {
	//	_, err = answer.StyledText(ctx, html.String(nil, StartMessage))
	//} else {
	//	_, err = answer.Text(ctx, "Вижу, Вы уже использовали бота, так что знаете что делать!"+
	//		"\nА если забыли, просто напишите /help")
	//}
	answer := c.Sender.Answer(e, upd)
	_, err = answer.StyledText(ctx, html.String(nil, StartMessage))
	return
}
