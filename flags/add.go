package flags

import (
	"fmt"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/util"
	"strings"
)

func Add(m *tg.Message, u *database.User) string {
	if strings.HasPrefix(m.Message, "/cancel") {
		database.DB.Model(&u).Select("Flag", "FlagData").Updates(database.User{Flag: "", FlagData: ""})
		return "Действие отменено"
	} else {
		if sticker, ok := util.StickerFromMedia(m.Media); ok {
			// if there is a sticker, check if there is such a tag attached to it,
			// if not — add one
			sTag := database.StickerTag{
				User:       u.UserID,
				StickerID:  sticker.ID,
				DocumentID: sticker.DocumentID,
				AccessHash: sticker.AccessHash,
				Tag:        u.FlagData,
			}
			answer, _ := sTag.CheckAndAdd()
			if !strings.HasPrefix(answer, "Что") {
				database.DB.Model(u).Select("Flag", "FlagData").Updates(database.User{Flag: "", FlagData: ""})
			}
			return answer
		} else {
			return fmt.Sprintf("В сообщении нет стикера! Отправьте мне стикер для "+
				"тега \"%s\" или отмените действие командой /cancel",
				u.FlagData)
		}
	}
}
