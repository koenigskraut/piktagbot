package commands

import (
	"context"
	"errors"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/flags"
	"github.com/koenigskraut/piktagbot/util"
)

func Check(ctx context.Context, e tg.Entities, upd *tg.UpdateNewMessage, c *HelperCapture, _ string) error {
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
			tags, err := (&database.StickerTag{User: user.UserID, StickerID: sticker.ID}).GetAllForUser()
			if err != nil {
				_, errMsg := answer.Text(ctx, "Что-то пошло не так, попробуйте ещё раз!")
				return errors.Join(err, errMsg)
			}
			if len(tags) == 0 {
				_, err := answer.Text(ctx, "У этого стикера нет ни одного тега!")
				return err
			}
			_, err = answer.StyledText(ctx, html.String(nil, util.CheckStickerResponse(tags)))
			return err
		}
		_, err := answer.Text(ctx, "В прикреплённом сообщении нет стикера! Отправьте мне стикер для просмотра тегов!")
		return err
	}

	// case 2: message is not a reply
	// set a waiting-for-a-sticker flag
	if err := user.SetFlag(flags.CheckTag, ""); err != nil {
		_, msgErr := answer.Text(ctx, "Что-то пошло не так, попробуйте ещё раз!")
		return errors.Join(err, msgErr)
	}
	_, err := answer.Text(ctx, "Теперь отправьте мне стикер, у которого нужно посмотреть тег(и)")
	return err
}
