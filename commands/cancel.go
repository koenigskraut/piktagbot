package commands

import (
	"context"
	"github.com/gotd/td/telegram/message"
)

func Cancel(ctx context.Context, answer *message.RequestBuilder) error {
	_, err := answer.Text(ctx, "Нечего отменять!")
	return err
}
