package flags

import (
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/util"
	"log"
	"strings"
)

func Check(m *tg.Message, u *database.User) string {
	if strings.HasPrefix(m.Message, "/cancel") {
		database.DB.Model(&u).Select("Flag", "FlagData").Updates(database.User{Flag: "", FlagData: ""})
		return "Действие отменено"
	} else {
		if sticker, ok := util.StickerFromMedia(m.Media); ok {
			tags, err := (&database.StickerTag{User: u.UserID, StickerID: sticker.ID}).GetAllForUser()
			if err != nil {
				log.Println(err)
				return "Что-то пошло не так, попробуйте ещё раз!"
			}
			if len(tags) == 0 {
				return "У этого стикера нет ни одного тега!"
			}
			return util.CheckStickerResponse(tags)
		} else {
			return "В сообщении нет стикера! Отправьте мне стикер для " +
				"просмотра тегов или отмените действие командой /cancel"
		}
	}
}
