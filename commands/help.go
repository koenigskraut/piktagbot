package commands

import (
	"context"
	_ "embed"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
)

//go:embed helpMessage.txt
var HelpMessage string

func Help(ctx context.Context, answer *message.RequestBuilder) (err error) {
	_, err = answer.StyledText(ctx, html.String(nil, HelpMessage))
	return
}
