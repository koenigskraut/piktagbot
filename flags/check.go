package flags

import (
	"context"
	"errors"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/util"
	"strings"
)

func Check(ctx context.Context, m *tg.Message, u *database.User, answer *message.RequestBuilder) error {
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
		_, err := answer.Text(ctx, "В сообщении нет стикера! Отправьте мне стикер для "+
			"просмотра тегов или отмените действие командой /cancel")
		return err
	}
	tags, err := (&database.StickerTag{User: u.UserID, StickerID: sticker.ID}).GetAllForUser()
	if err != nil {
		_, errMsg := answer.Text(ctx, "Что-то пошло не так, попробуйте ещё раз!")
		return errors.Join(err, errMsg)
	}
	if len(tags) == 0 {
		_, err := answer.Text(ctx, "У этого стикера нет ни одного тега!")
		return err
	}
	u.Flag, u.FlagData = NoFlag, ""
	if err := u.Save(); err != nil {
		_, errMsg := answer.Text(ctx, "Что-то пошло не так, попробуйте ещё раз!")
		return errors.Join(err, errMsg)
	}
	_, err = answer.StyledText(ctx, html.String(nil, util.CheckStickerResponse(tags)))
	return err
}
