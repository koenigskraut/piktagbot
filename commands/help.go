package commands

import (
	"context"
	_ "embed"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/tg"
)

//go:embed helpMessage.txt
var HelpMessage string

func Help(ctx context.Context, e tg.Entities, upd *tg.UpdateNewMessage, c *HelperCapture, _ string) error {
	_, err := c.Sender.Answer(e, upd).StyledText(ctx, html.String(nil, HelpMessage))
	return err
}
