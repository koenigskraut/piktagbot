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
	"net/url"
	"strconv"
	"time"
)

func handleInline(client *tg.Client) func(context.Context, tg.Entities, *tg.UpdateBotInlineQuery) error {
	sender := message.NewSender(client)
	return func(ctx context.Context, entities tg.Entities, update *tg.UpdateBotInlineQuery) (err error) {
		var q []*db.StickerTag

		u := db.User{UserID: update.UserID}
		if _, e := u.Get(); e != nil {
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
		webAppUser := &webapp.User{}
		for _, e := range entities.Users {
			if e.ID == update.UserID {
				webAppUser.FillFrom(e)
				break
			}
		}
		webAppParams := webapp.InitData{
			&webapp.QueryID{Data: strconv.FormatInt(update.QueryID, 10)}, webAppUser,
			&webapp.AuthDate{Data: time.Now().Unix()}, &webapp.Prefix{Data: update.Query},
		}
		signed, err := webAppParams.Sign(util.GetSecretKey())
		if err != nil {
			return err
		}
		URL := fmt.Sprintf("https://koenigskraut.ru:55506?%s", url.PathEscape(string(signed)))

		if len(as) > 0 {
			if len(as) > 50 {
				as = as[:50]
			}
			if _, err := w.SwitchWebview("Изменить порядок стикеров", URL).Set(ctx, as...); err != nil {
				return err
			}
		} else {
			_, err := w.SwitchPM("Начать создавать теги!", "a").Set(ctx)
			return err
		}
		return nil
	}
}
