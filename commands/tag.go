// func: RunTag; triggers: tag; response: string;

package commands

import (
	"context"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/util"
	"strings"
)

func Tag(ctx context.Context, answer *message.RequestBuilder, m *tg.Message, client *tg.Client, clear string) error {
	userID := m.PeerID.(*tg.PeerUser).UserID
	text := strings.TrimSpace(clear)

	// if there is no tag in a message return immediately
	if text == "" {
		_, err := answer.Text(ctx, "Вы не написали тег, который хотите прикрепить к стикеру!")
		return err
	}

	// case 1: message is a reply, handle re message
	if m.ReplyTo != nil {
		// get re message
		mRep, _ := client.MessagesGetMessages(
			ctx,
			[]tg.InputMessageClass{&tg.InputMessageReplyTo{ID: m.ID}},
		)
		// check if there is a sticker in it
		media := mRep.(*tg.MessagesMessages).Messages[0].(*tg.Message).Media
		if doc, ok := util.DocFromMedia(media); ok {
			// if true, check if it has such a tag, add one if not
			sTag := database.StickerTag{User: userID, DocumentID: doc.ID, AccessHash: doc.AccessHash, Tag: text}
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
	user := database.User{UserID: userID}
	if _, err := user.Get(); err != nil {
		answer.Text(ctx, "Что-то пошло не так, попробуйте ещё раз!")
		return err
	}
	if err := user.SetFlag("add-sticker", text); err != nil {
		answer.Text(ctx, "Что-то пошло не так, попробуйте ещё раз!")
		return err
	}
	_, err := answer.Text(ctx, "Теперь отправьте мне стикер, к которому нужно прикрепить тег")
	return err
}
