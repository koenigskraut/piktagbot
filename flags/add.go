package flags

import (
	"context"
	"errors"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/util"
	"strings"
)

func Add(ctx context.Context, m *tg.Message, u *database.User, answer *message.RequestBuilder) error {
	if strings.HasPrefix(m.Message, "/cancel") {
		u.Flag, u.FlagData = NoFlag, ""
		if err := u.Save(); err != nil {
			return err
		}
		_, err := answer.Text(ctx, "Действие отменено")
		return err
	}
	sticker, ok := util.StickerFromMedia(m.Media)
	if !ok {
		_, err := answer.Text(ctx, "В сообщении нет стикера! Отправьте мне стикер для удаления тегов или "+
			"отмените действие командой /cancel")
		return err
	}
	// if there is a sticker, check if there is such a tag attached to it,
	// if not — add one
	sTag := database.StickerTag{
		User:      u.UserID,
		StickerID: sticker.ID,
		Tag:       u.FlagData,
	}
	text, errDB := sTag.CheckAndAdd()
	if errDB == nil {
		u.Flag, u.FlagData = NoFlag, ""
		if err := u.Save(); err != nil {
			return err
		}
	}
	_, errMsg := answer.Text(ctx, text)
	return errors.Join(errDB, errMsg)
}
