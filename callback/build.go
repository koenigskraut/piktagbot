package callback

// TODO review all of this file, first glance: not great

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gotd/td/telegram/message/markup"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/database"
)

const maxElementsOnPage = 10

const (
	ActionNone = iota
	ActionRemove
	ActionBackward
	ActionForward
	ActionBegin
	ActionEnd
	ActionDone
)

const (
	TextNone     = "⏺"
	TextBegin    = "⏮"
	TextBackward = "⬅"
	TextForward  = "➡"
	TextEnd      = "⏭"
	TextDone     = "Готово"
)

var MarkupError = errors.New("markup build: no tags found")

func prepareTagID(tagID, docID uint64, page uint16) []byte {
	byteTagID := make([]byte, 8)
	binary.LittleEndian.PutUint64(byteTagID, tagID)
	byteDocID := make([]byte, 8)
	binary.LittleEndian.PutUint64(byteDocID, docID)
	bytePage := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytePage, page)
	b := append([]byte{ActionRemove}, byteTagID...)
	b = append(b, byteDocID...)
	b = append(b, bytePage...)
	return b
}

func preparePage(code int, docID uint64, page uint16) []byte {
	byteDocID := make([]byte, 8)
	binary.LittleEndian.PutUint64(byteDocID, docID)
	bytePage := make([]byte, 2)
	binary.LittleEndian.PutUint16(bytePage, page)
	b := append([]byte{byte(code)}, byteDocID...)
	b = append(b, bytePage...)
	return b
}

// BuildMarkup renders buttons for tag deletion for a given
// sticker (docID) and user (userID), stopping on a given page
func BuildMarkup(stickerID uint64, userID int64, page uint16) (tg.ReplyMarkupClass, error) {
	// get tags for a given sticker
	tags, err := (&database.StickerTag{User: userID, StickerID: stickerID}).GetAllForUser()
	//tags, err := getStickerTags(docID, userID)
	if err != nil {
		return nil, err
	}
	// and their number, if it is zero, return
	tagLen := len(tags)
	if tagLen == 0 {
		return nil, MarkupError
	}
	const tagsOnPage = 7
	pages := (tagLen - 1) / tagsOnPage
	// if we are on a page that no longer exists, fix it
	if int(page) > pages {
		page = uint16(pages)
	}
	var rows []tg.KeyboardButtonRow
	// if we don't need nav buttons
	if tagLen < 11 {
		rows = make([]tg.KeyboardButtonRow, tagLen)
		for i := range tags {
			b := prepareTagID(tags[i].ID, stickerID, 0)
			button := markup.Callback(tags[i].Tag, b)
			rows[i] = markup.Row(button)
		}
		return markup.InlineKeyboard(rows...), nil
	}
	// if we do need them
	relevantTags := tags[page*tagsOnPage:]
	if len(relevantTags) > tagsOnPage {
		relevantTags = relevantTags[:tagsOnPage]
	}
	// can't think of anything more elegant, could use map or slice,
	// but it's not worth it
	var first, backward, forward, last tg.KeyboardButtonClass

	if page > 1 {
		first = markup.Callback(TextBegin, preparePage(ActionBegin, stickerID, page))
		backward = markup.Callback(TextBackward, preparePage(ActionBackward, stickerID, page))
	} else {
		first = markup.Callback(TextNone, []byte{ActionNone})
		if page > 0 {
			backward = markup.Callback(TextBackward, preparePage(ActionBackward, stickerID, page))
		} else {
			backward = markup.Callback(TextNone, []byte{ActionNone})
		}
	}

	if pages-int(page) > 1 {
		last = markup.Callback(TextEnd, preparePage(ActionEnd, stickerID, page))
		forward = markup.Callback(TextForward, preparePage(ActionForward, stickerID, page))
	} else {
		last = markup.Callback(TextNone, []byte{ActionNone})
		if pages > int(page) {
			forward = markup.Callback(TextForward, preparePage(ActionForward, stickerID, page))
		} else {
			forward = markup.Callback(TextNone, []byte{ActionNone})
		}
	}

	navButtons := markup.Row(
		first, backward,
		markup.Callback(fmt.Sprintf("%d / %d", page+1, pages+1), []byte{ActionNone}),
		forward, last,
	)
	doneRow := markup.Row(markup.Callback(TextDone, []byte{ActionDone}))

	rows = make([]tg.KeyboardButtonRow, 0, (maxElementsOnPage-tagsOnPage)+len(relevantTags))
	rows = append(rows, navButtons)
	for _, t := range relevantTags {
		b := prepareTagID(t.ID, stickerID, page)
		button := markup.Callback(t.Tag, b)
		rows = append(rows, markup.Row(button))
	}
	rows = append(rows, navButtons, doneRow)
	return markup.InlineKeyboard(rows...), nil
}
