// func: RunTag; triggers: tag; response: string;

package commands

import (
	"context"
	"errors"
	"github.com/gotd/td/tg"
	db "github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/util"
	"strings"
)

func Tag(ctx context.Context, e tg.Entities, upd *tg.UpdateNewMessage, c *HelperCapture, clear string) (err error) {
	m := upd.Message.(*tg.Message)
	userID := m.PeerID.(*tg.PeerUser).UserID
	user := c.UserCapture.(*MessageSemaphore).GetCurrentLock(userID).DBUser
	answer := c.Sender.Answer(e, upd)
	text := strings.TrimSpace(clear)

	// if there is no tag in a message return immediately
	if text == "" {
		_, err := answer.Text(ctx, "Вы не написали тег, который хотите прикрепить к стикеру!")
		return err
	}

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
			// if true, check if it has such a tag, add one if not
			sTag := db.StickerTag{
				User:      userID,
				StickerID: sticker.ID,
				Tag:       text,
			}
			resp, _ := sTag.CheckAndAdd()
			_, err := answer.Text(ctx, resp)
			return err
		}
		_, err := answer.Textf(
			ctx,
			"В прикреплённом сообщении нет стикера! Отправьте мне стикер для тега \"%s\"",
			text,
		)
		return err
	}

	// case 2: message is not a reply
	// remember the tag and set a waiting-for-a-sticker flag
	if err := user.SetFlag("add-sticker", text); err != nil {
		_, msgErr := answer.Text(ctx, "Что-то пошло не так, попробуйте ещё раз!")
		return errors.Join(err, msgErr)
	}
	_, err = answer.Text(ctx, "Теперь отправьте мне стикер, к которому нужно прикрепить тег")
	return err
}
