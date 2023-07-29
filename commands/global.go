// func: RunGlobal; triggers: global; response: string;

package commands

import (
	"context"
	"github.com/gotd/td/telegram/message"
	db "github.com/koenigskraut/piktagbot/database"
)

func Global(ctx context.Context, answer *message.RequestBuilder, user *db.User) (err error) {
	// can't update DB — notify user
	if errDB := user.SwitchGlobal(); errDB != nil {
		answer.Text(ctx, "Произошла какая-то ошибка, попробуйте ещё раз!")
		return errDB
	}

	// no error, send user status
	if user.GlobalTag {
		_, err = answer.Text(ctx, "Глобальные теги включены")
	} else {
		_, err = answer.Text(ctx, "Глобальные теги выключены")
	}
	return
}
