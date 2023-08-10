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

const (
	finish = iota
	errNoSticker
	errExists
)

const (
	finishCancel = "Действие отменено"
	finishDone   = "Действие завершено"

	errNoStickerOne  = "В сообщении нет стикера! Отправьте мне стикер для добавления тегов или отмените действие командой /cancel"
	errNoStickerMany = "В сообщении нет стикера! Отправьте мне стикеры для добавления тегов или завершите действие командой /done"

	errGeneral = "Что-то пошло не так, попробуйте ещё раз"

	errExistsOne  = "У этого стикера уже есть этот тег, действие отменено"
	errExistsMany = "У этого стикера уже есть этот тег, отправьте другие!"

	successOne  = "Тег добавлен!"
	successMany = "Тег добавлен! Пришлите мне следующий стикер или завершите действие командой /done"
)

type sm map[bool]string

func handleAdd(isOne bool) FlagHandler {
	chooseStr := map[uint8]map[bool]string{
		finish:       sm{true: finishDone, false: finishCancel},
		errNoSticker: sm{true: errNoStickerOne, false: errNoStickerMany},
		errExists:    sm{true: errExistsOne, false: errExistsMany},
	}
	//flag := map[bool]int8{true: flags.AddTag, false: flags.AddTags}
	return func(ctx context.Context, m *tg.Message, u *database.User, answer *message.RequestBuilder) error {
		done := strings.HasPrefix(m.Message, "/done")
		cancel := strings.HasPrefix(m.Message, "/cancel")
		if done || cancel {
			u.Flag, u.FlagData = NoFlag, ""
			if err := u.Save(); err != nil {
				return err
			}
			_, err := answer.Text(ctx, chooseStr[finish][done])
			return err
		}
		sticker, ok := util.StickerFromMedia(m.Media)
		if !ok {
			_, err := answer.Text(ctx, chooseStr[errNoSticker][isOne])
			return err
		}
		// if there is a sticker, check if there is such a tag attached to it,
		// if not — add one
		sTag := database.StickerTag{
			User:      u.UserID,
			StickerID: sticker.ID,
			Tag:       u.FlagData,
		}
		errDB := sTag.CheckAndAdd()
		if errDB != nil {
			if !errors.Is(errDB, database.StickerTagExists) {
				_, errMsg := answer.Text(ctx, errGeneral)
				return errors.Join(errDB, errMsg)
			}
			if isOne {
				u.Flag, u.FlagData = NoFlag, ""
				if err := u.Save(); err != nil {
					_, errMsg := answer.Text(ctx, errGeneral)
					return errors.Join(err, errMsg)
				}
			}
			_, errMsg := answer.Text(ctx, chooseStr[errExists][isOne])
			return errMsg
		}
		if !isOne {
			_, err := answer.Text(ctx, successMany)
			return err
		}
		u.Flag, u.FlagData = NoFlag, ""
		if err := u.Save(); err != nil {
			_, errMsg := answer.Text(ctx, errGeneral)
			return errors.Join(err, errMsg)
		}
		_, err := answer.Text(ctx, successOne)
		return err
	}
}

var AddOne = handleAdd(true)
var AddMany = handleAdd(false)
