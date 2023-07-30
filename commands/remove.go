package commands

import (
	"context"
	"errors"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/callback"
	"github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/util"
)

func Remove(ctx context.Context, answer *message.RequestBuilder, m *tg.Message, client *tg.Client) error {
	userID := m.PeerID.(*tg.PeerUser).UserID

	// case 1: message is a reply, handle re message
	if m.ReplyTo != nil {
		// get re message
		mRep, _ := client.MessagesGetMessages(
			ctx,
			[]tg.InputMessageClass{&tg.InputMessageReplyTo{ID: m.ID}},
		)
		// check if there is a sticker in it
		media := mRep.(*tg.MessagesMessages).Messages[0].(*tg.Message).Media
		if sticker, ok := util.StickerFromMedia(media); ok {
			// if true, send message with tag deletion buttons
			markup, err := callback.BuildMarkup(sticker.ID, userID, 0)
			if err != nil {
				if errors.Is(err, callback.MarkupError) {
					answer.Text(ctx, "У этого стикера нет ни одного тега!")
				} else {
					answer.Text(ctx, "Что-то пошло не так, попробуйте ещё раз!")
				}
				return err
			}
			_, err = answer.Markup(markup).Text(ctx, "Выберите тег для удаления:")
			return err
		}
		_, err := answer.Text(ctx, "В прикреплённом сообщении нет стикера! Отправьте мне стикер для удаления тегов!")
		return err
	}

	// case 2: message is not a reply
	// set a waiting-for-a-sticker flag
	user := database.User{UserID: userID}
	if _, err := user.Get(); err != nil {
		answer.Text(ctx, "Что-то пошло не так, попробуйте ещё раз!")
		return err
	}
	if err := user.SetFlag("remove-tag", ""); err != nil {
		answer.Text(ctx, "Что-то пошло не так, попробуйте ещё раз!")
		return err
	}
	_, err := answer.Text(ctx, "Теперь отправьте мне стикер, у которого нужно удалить тег(и)")
	return err
}
