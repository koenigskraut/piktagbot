package flags

import (
	"errors"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/callback"
	"github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/util"
	"log"
	"strings"
)

func Remove(m *tg.Message, u *database.User) (string, tg.ReplyMarkupClass) {
	if strings.HasPrefix(m.Message, "/cancel") {
		database.DB.Model(&u).Select("Flag", "FlagData").Updates(database.User{Flag: "", FlagData: ""})
		return "Действие отменено", nil
	} else {
		if sticker, ok := util.StickerFromMedia(m.Media); ok {
			// if there is a sticker, let's build a keyboard for it with
			// tags to delete
			markup, err := callback.BuildMarkup(sticker.ID, u.UserID, 0)
			if err != nil {
				if errors.Is(err, callback.MarkupError) {
					return "У этого стикера нет ни одного тега!", nil
				}
				log.Println(err)
				return "Неизвестная ошибка", nil
			}
			database.DB.Model(u).Select("Flag", "FlagData").Updates(database.User{Flag: "", FlagData: ""})
			return "Выберите тег для удаления:", markup
		} else {
			return "В сообщении нет стикера! Отправьте мне стикер для " +
				"удаления тегов или отмените действие командой /cancel", nil
		}
	}
}
