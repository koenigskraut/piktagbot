package main

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/inline"
	"github.com/gotd/td/tg"
	db "github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/util"
	"github.com/koenigskraut/piktagbot/webapp"
	"strconv"
	"time"
)

func handleInline(client *tg.Client) func(context.Context, tg.Entities, *tg.UpdateBotInlineQuery) error {
	sender := message.NewSender(client)
	return func(ctx context.Context, entities tg.Entities, update *tg.UpdateBotInlineQuery) (err error) {
		var q []*db.StickerTag

		const limit = 50
		offset, _ := strconv.ParseInt(update.Offset, 10, 32)
		nextOffset := ""

		u := db.User{UserID: update.UserID}
		if e := u.Get(); e != nil {
			return e
		}
		if update.Query != "" {
			q, err = u.SearchStickers(update.Query)
		} else {
			q, err = u.RecentStickers()
		}
		if err != nil {
			return err
		}
		q = q[offset:]
		if len(q) > limit {
			q = q[:limit]
			nextOffset = strconv.FormatInt(offset+limit, 10)
		}

		as := make([]inline.ResultOption, len(q))
		for i, st := range q {
			as[i] = inline.Sticker(
				&tg.InputDocument{
					ID:         st.Sticker.DocumentID,
					AccessHash: st.Sticker.AccessHash,
				}, inline.MediaAuto(""),
			)
		}

		w := sender.Inline(update).CacheTimeSeconds(0).Private(true).NextOffset(nextOffset).Gallery(true)
		webAppUser := &webapp.User{}
		for _, e := range entities.Users {
			if e.ID == update.UserID {
				webAppUser.FillFrom(e)
				break
			}
		}
		webAppParams := webapp.InitDataList{
			&webapp.QueryID{Data: strconv.FormatInt(update.QueryID, 10)}, webAppUser,
			&webapp.AuthDate{Data: time.Now().Unix()}, &webapp.Prefix{Data: update.Query},
		}
		if err := webAppParams.Sign(util.GetSecretKey()); err != nil {
			return err
		}
		signed, err := webAppParams.Serialize('&')
		if err != nil {
			return err
		}
		URL := fmt.Sprintf("https://%s:%s?%s", appDomain, appPort, string(signed))
		switch len(as) {
		case 0:
			_, err = w.SwitchPM("Начать создавать теги!", "a").Set(ctx)
		case 1:
			_, err = w.SwitchPM("Добавить теги!", "a").Set(ctx)
		default:
			_, err = w.SwitchWebview("Изменить порядок стикеров", URL).Set(ctx, as...)
		}
		return err
	}
}
