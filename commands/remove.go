package commands

import (
	"context"
	"errors"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/callback"
	"github.com/koenigskraut/piktagbot/flags"
	"github.com/koenigskraut/piktagbot/util"
)

func Remove(ctx context.Context, e tg.Entities, upd *tg.UpdateNewMessage, c *HelperCapture, _ string) (err error) {
	m, user := c.UserCapture.(*MessageSemaphore).MessageUserFromUpdate(upd)
	answer := c.Sender.Answer(e, upd)

	// case 1: message is a reply, handle re message
	if m.ReplyTo != nil {
		// get re message
		mRep, _ := c.Client.MessagesGetMessages(
			ctx,
			[]tg.InputMessageClass{&tg.InputMessageReplyTo{ID: m.ID}},
		)
		// check if there is a sticker in it
		media := mRep.(*tg.MessagesMessages).Messages[0].(*tg.Message).Media
		if sticker, ok := util.StickerFromMedia(media); ok {
			// if true, send message with tag deletion buttons
			markup, err := callback.BuildMarkup(sticker.ID, user.UserID, 0)
			if err != nil {
				var msgErr error
				if errors.Is(err, callback.MarkupError) {
					_, msgErr = answer.Text(ctx, "У этого стикера нет ни одного тега!")
				} else {
					_, msgErr = answer.Text(ctx, "Что-то пошло не так, попробуйте ещё раз!")
				}
				return errors.Join(err, msgErr)
			}
			_, err = answer.Markup(markup).Text(ctx, "Выберите тег для удаления:")
			return err
		}
		_, err := answer.Text(ctx, "В прикреплённом сообщении нет стикера! Отправьте мне стикер для удаления тегов!")
		return err
	}

	// case 2: message is not a reply
	// set a waiting-for-a-sticker flag
	if err := user.SetFlag(flags.RemoveTag, ""); err != nil {
		_, msgErr := answer.Text(ctx, "Что-то пошло не так, попробуйте ещё раз!")
		return errors.Join(err, msgErr)
	}
	_, err = answer.Text(ctx, "Теперь отправьте мне стикер, у которого нужно удалить тег(и)")
	return err
}
