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
		var as []inline.ResultOption
		var q []*db.StickerTag

		u := db.User{UserID: update.UserID}
		if _, e := u.Get(); e != nil {
			return e
		}
		checkUnique := make(map[int64]struct{})
		if update.Query != "" {
			q, _ = u.SearchStickers(update.Query)
		} else {
			q, _ = u.RecentStickers()
		}

		for _, i := range q {
			if _, ok := checkUnique[i.DocumentID]; !ok {
				as = append(as, inline.Sticker(
					&tg.InputDocument{
						ID:         i.DocumentID,
						AccessHash: i.AccessHash,
					}, inline.MediaAuto(""),
				))
				checkUnique[i.DocumentID] = struct{}{}
			}
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
