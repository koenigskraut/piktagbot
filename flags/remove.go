package flags

import (
	"context"
	"errors"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/callback"
	"github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/util"
	"strings"
)

func Remove(ctx context.Context, m *tg.Message, u *database.User, answer *message.RequestBuilder) error {
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
	// if there is a sticker, let's build a keyboard for it with tags to delete
	markup, errMarkup := callback.BuildMarkup(sticker.ID, u.UserID, 0)
	if errMarkup != nil {
		msg := "Неизвестная ошибка"
		if errors.Is(errMarkup, callback.MarkupError) {
			msg = "У этого стикера нет ни одного тега!"
		}
		_, errMsg := answer.Text(ctx, msg)
		return errors.Join(errMarkup, errMsg)
	}
	u.Flag, u.FlagData = NoFlag, ""
	if err := u.Save(); err != nil {
		return err
	}
	_, err := answer.Markup(markup).Text(ctx, "Выберите тег для удаления:")
	return err
}
