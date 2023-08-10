// func: RunTag; triggers: tag; response: string;

package commands

import (
	"context"
	"errors"
	"fmt"
	"github.com/gotd/td/tg"
	db "github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/flags"
	"github.com/koenigskraut/piktagbot/util"
	"strings"
)

const (
	errNoTag = iota
	errNoSticker
	errExists
	successFlag
	successRe
)

const (
	errNoTagOne  = "Вы не написали тег, который хотите прикрепить к стикеру!"
	errNoTagMany = "Вы не написали тег, который хотите прикрепить к стикерам!"

	errNoStickerOne  = `В прикреплённом сообщении нет стикера, действие отменено%.s`
	errNoStickerMany = `В прикреплённом сообщении нет стикера! Отправьте мне стикеры для тега "%s"`

	errGeneral = "Что-то пошло не так, попробуйте ещё раз!"

	errExistsOne  = "У этого стикера уже есть этот тег, действие отменено"
	errExistsMany = "У этого стикера уже есть этот тег, отправьте другие!"

	successFlagOne  = "Теперь отправьте мне стикер, к которому нужно прикрепить тег"
	successFlagMany = "Теперь отправьте мне стикеры, к которым нужно прикрепить теги"

	successReOne  = "Тег добавлен!"
	successReMany = "Тег добавлен! Пришлите мне следующий стикер или завершите действие командой /done"
)

type sm map[bool]string

func handleTag(isOne bool) CommandHandler {
	chooseStr := map[uint8]map[bool]string{
		errNoTag:     sm{true: errNoTagOne, false: errNoTagMany},
		errNoSticker: sm{true: errNoStickerOne, false: errNoStickerMany},
		errExists:    sm{true: errExistsOne, false: errExistsMany},
		successFlag:  sm{true: successFlagOne, false: successFlagMany},
		successRe:    sm{true: successReOne, false: successReMany},
	}
	flag := map[bool]int8{true: flags.AddTag, false: flags.AddTags}
	return func(ctx context.Context, e tg.Entities, upd *tg.UpdateNewMessage, c *HelperCapture, clear string) (err error) {
		m, user := c.UserCapture.(*MessageSemaphore).MessageUserFromUpdate(upd)
		answer := c.Sender.Answer(e, upd)
		text := strings.TrimSpace(clear)

		// if there is no tag in a message return immediately
		if text == "" {
			_, err := answer.Text(ctx, chooseStr[errNoTag][isOne])
			return err
		}

		// case 1: message is not a reply
		// remember the tag and set a waiting-for-a-sticker flag
		if m.ReplyTo == nil {
			if err := user.SetFlag(flag[isOne], text); err != nil {
				_, msgErr := answer.Text(ctx, errGeneral)
				return errors.Join(err, msgErr)
			}
			_, err = answer.Text(ctx, chooseStr[successFlag][isOne])
			return err
		}

		// case 2: message is a reply, handle re message
		// many stickers wanted? set flag immediately
		if !isOne {
			if err := user.SetFlag(flag[isOne], text); err != nil {
				_, msgErr := answer.Text(ctx, errGeneral)
				return errors.Join(err, msgErr)
			}
		}
		// get re message
		mRep, _ := c.Client.MessagesGetMessages(
			ctx,
			[]tg.InputMessageClass{&tg.InputMessageReplyTo{ID: m.ID}},
		)
		// check if there is a sticker in it
		media := mRep.(*tg.MessagesMessages).Messages[0].(*tg.Message).Media
		sticker, ok := util.StickerFromMedia(media)
		// if false return
		if !ok {
			fmt.Println()
			_, err := answer.Textf(
				ctx,
				chooseStr[errNoSticker][isOne],
				text,
			)
			return err
		}
		// if true, check if sticker has such a tag, add one if not
		sTag := db.StickerTag{
			User:      user.UserID,
			StickerID: sticker.ID,
			Tag:       text,
		}
		err = sTag.CheckAndAdd()
		var errTwo error
		switch err {
		case nil:
			_, errTwo = answer.Text(ctx, chooseStr[successRe][isOne])
		case db.StickerTagExists:
			err = nil
			_, errTwo = answer.Text(ctx, chooseStr[errExists][isOne])
		default:
			_, errTwo = answer.Text(ctx, errGeneral)
		}
		return errors.Join(err, errTwo)
	}
}

var Tag = handleTag(true)
var Tags = handleTag(false)
