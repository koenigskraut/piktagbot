package commands

import (
	"context"
	"github.com/gotd/td/tg"
)

func Cancel(ctx context.Context, e tg.Entities, upd *tg.UpdateNewMessage, c *HelperCapture, _ string) error {
	_, err := c.Sender.Answer(e, upd).Text(ctx, "Нечего отменять!")
	return err
}
