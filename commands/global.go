// func: RunGlobal; triggers: global; response: string;

package commands

import (
	"context"
	"errors"
	"github.com/gotd/td/tg"
)

func Global(ctx context.Context, e tg.Entities, upd *tg.UpdateNewMessage, c *HelperCapture, _ string) (err error) {
	_, user := c.UserCapture.(*MessageSemaphore).MessageUserFromUpdate(upd)
	answer := c.Sender.Answer(e, upd)
	// can't update DB — notify user
	if errDB := user.SwitchGlobal(); errDB != nil {
		_, msgErr := answer.Text(ctx, "Произошла какая-то ошибка, попробуйте ещё раз!")
		return errors.Join(errDB, msgErr)
	}

	// no error, send user status
	if user.GlobalTag {
		_, err = answer.Text(ctx, "Глобальные теги включены")
	} else {
		_, err = answer.Text(ctx, "Глобальные теги выключены")
	}
	return
}
