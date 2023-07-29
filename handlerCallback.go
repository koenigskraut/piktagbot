package main

import (
	"context"
	"encoding/binary"
	"errors"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/callback"
	db "github.com/koenigskraut/piktagbot/database"
	//	"gorm.io/gorm"
	//	"sync"
)

// TODO rework this monstrosity
func handleCallback(client *tg.Client) func(context.Context, tg.Entities, *tg.UpdateBotCallbackQuery) error {
	return func(ctx context.Context, e tg.Entities, update *tg.UpdateBotCallbackQuery) (err error) {
		var message string

		if update.Data[0] == callback.ActionNone {
			return answerCallback(ctx, update.QueryID, message, client)
		}

		page := binary.LittleEndian.Uint16(update.Data[len(update.Data)-2:])
		var docID int64
		if update.Data[0] == callback.ActionRemove {
			docID = int64(binary.LittleEndian.Uint64(update.Data[9:17]))
		} else {
			docID = int64(binary.LittleEndian.Uint64(update.Data[1:9]))
		}
		userID := update.UserID
		var markup *tg.ReplyInlineMarkup

		switch update.Data[0] {
		case callback.ActionRemove:
			tagID := binary.LittleEndian.Uint64(update.Data[1:9])
			if errDelete := db.DB.Delete((&db.StickerTag{}), tagID).Error; errDelete != nil {
				message = "Что-то пошло не так!"
			} else {
				message = "Тег удалён!"
				markup, err = callback.BuildMarkup(docID, userID, page)
				if err != nil {
					if errors.Is(err, callback.MarkupError) {
						message = "Тегов для этого стикера больше нет!"
						err = answerCallback(ctx, update.QueryID, message, client)
						_, err = client.MessagesEditMessage(ctx, &tg.MessagesEditMessageRequest{
							Peer:    &tg.InputPeerUser{UserID: update.UserID},
							ID:      update.MsgID,
							Message: "Все теги у выбранного стикера удалены!",
						})
						return err
					} else {
						message = "Что-то пошло не так!"
					}
					err = answerCallback(ctx, update.QueryID, message, client)
					return err
				}
			}
		case callback.ActionBegin:
			markup, err = callback.BuildMarkup(docID, userID, 0)
			if err != nil {
				message = "Что-то пошло не так!"
				err = answerCallback(ctx, update.QueryID, message, client)
				return err
			}
		case callback.ActionEnd:
			markup, err = callback.BuildMarkup(docID, userID, 65535)
			if err != nil {
				message = "Что-то пошло не так!"
				err = answerCallback(ctx, update.QueryID, message, client)
				return err
			}
		case callback.ActionBackward:
			page--
			markup, err = callback.BuildMarkup(docID, userID, page)
			if err != nil {
				message = "Что-то пошло не так!"
				err = answerCallback(ctx, update.QueryID, message, client)
				return err
			}
		case callback.ActionForward:
			page++
			markup, err = callback.BuildMarkup(docID, userID, page)
			if err != nil {
				message = "Что-то пошло не так!"
				err = answerCallback(ctx, update.QueryID, message, client)
				return err
			}
		}

		_, err = client.MessagesEditMessage(ctx, &tg.MessagesEditMessageRequest{
			Peer:        &tg.InputPeerUser{UserID: update.UserID},
			ID:          update.MsgID,
			Message:     "Выберите тег для удаления:",
			ReplyMarkup: markup,
		})
		if err != nil {
			message = "Что-то пошло не так!"
		}
		return answerCallback(ctx, update.QueryID, message, client)
	}
}

// answerCallback is a simple wrapper that discards bool return value
func answerCallback(ctx context.Context, queryID int64, message string, client *tg.Client) error {
	_, err := client.MessagesSetBotCallbackAnswer(ctx, &tg.MessagesSetBotCallbackAnswerRequest{
		QueryID: queryID,
		Message: message,
	})
	return err
}
