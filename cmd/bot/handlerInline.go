package main

import (
	"context"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/inline"
	"github.com/gotd/td/tg"
	db "github.com/koenigskraut/piktagbot/database"
)

func handleInline(client *tg.Client) func(context.Context, tg.Entities, *tg.UpdateBotInlineQuery) error {
	sender := message.NewSender(client)
	return func(ctx context.Context, entities tg.Entities, update *tg.UpdateBotInlineQuery) error {
		var q []*db.StickerTag

		u := db.User{UserID: update.UserID}
		if _, e := u.Get(); e != nil {
			return e
		}
		if update.Query != "" {
			q, _ = u.SearchStickers(update.Query)
		} else {
			q, _ = u.RecentStickers()
		}
		_ = q

		as := make([]inline.ResultOption, len(q))
		for i, st := range q {
			as[i] = inline.Sticker(
				&tg.InputDocument{
					ID:         st.Sticker.DocumentID,
					AccessHash: st.Sticker.AccessHash,
				}, inline.MediaAuto(""),
			)
		}

		w := sender.Inline(update).CacheTimeSeconds(0).NextOffset("").Gallery(true)
		if len(as) > 0 {
			w.SwitchPM("Добавить новые теги", "n").Set(ctx, as...)
		} else {
			w.SwitchPM("Начать создавать теги!", "a").Set(ctx, inline.Article("", inline.MessageText("")))
		}
		return nil
	}
}
