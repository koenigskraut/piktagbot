package commands

import (
	"context"
	_ "embed"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
)

//go:embed startMessage.txt
var StartMessage string

func Start(ctx context.Context, answer *message.RequestBuilder, isNew bool) (err error) {
	if isNew {
		_, err = answer.StyledText(ctx, html.String(nil, StartMessage))
	} else {
		_, err = answer.Text(ctx, "Вижу, Вы уже использовали бота, так что знаете что делать!"+
			"\nА если забыли, просто напишите /help")
	}
	return
}
