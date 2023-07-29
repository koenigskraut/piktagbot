package callback

// TODO review all of this file, first glance: not great

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/database"
)

const (
	ActionNone = iota
	ActionRemove
	ActionBackward
	ActionForward
	ActionBegin
	ActionEnd
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
func BuildMarkup(docID, userID int64, page uint16) (*tg.ReplyInlineMarkup, error) {
	// get tags for a given sticker
	tags, err := (&database.StickerTag{User: userID, DocumentID: docID}).GetAllForUser()
	//tags, err := getStickerTags(docID, userID)
	if err != nil {
		return nil, err
	}
	// and their number, if it is zero, return
	tagLen := len(tags)
	if tagLen == 0 {
		return nil, MarkupError
	}
	markup := &tg.ReplyInlineMarkup{}
	pages := (tagLen - 1) / 8
	// if we are on a page that no longer exists, fix it
	if int(page) > pages {
		page = uint16(pages)
	}
	// if we don't need nav buttons
	if tagLen < 11 {
		markup.Rows = make([]tg.KeyboardButtonRow, tagLen)
		for i := range tags {
			b := prepareTagID(tags[i].ID, uint64(docID), 0)
			button := &tg.KeyboardButtonCallback{Text: tags[i].Tag, Data: b}
			markup.Rows[i] = tg.KeyboardButtonRow{Buttons: []tg.KeyboardButtonClass{button}}
		}
	} else {
		// if we do need them
		relevantTags := tags[page*8:]
		if len(relevantTags) > 8 {
			relevantTags = relevantTags[:8]
		}

		// can't think of anything more elegant, could use map or slice,
		// but it's not worth it
		var first, backward, forward, last tg.KeyboardButtonClass

		if page > 1 {
			first = &tg.KeyboardButtonCallback{Text: "⏮", Data: preparePage(ActionBegin, uint64(docID), page)}
			backward = &tg.KeyboardButtonCallback{Text: "⬅️", Data: preparePage(ActionBackward, uint64(docID), page)}
		} else {
			first = &tg.KeyboardButtonCallback{Text: "⏺", Data: []byte{ActionNone}}
			if page > 0 {
				backward = &tg.KeyboardButtonCallback{Text: "⬅️", Data: preparePage(ActionBackward, uint64(docID), page)}
			} else {
				backward = &tg.KeyboardButtonCallback{Text: "⏺", Data: []byte{ActionNone}}
			}
		}

		if pages-int(page) > 1 {
			last = &tg.KeyboardButtonCallback{Text: "⏭", Data: preparePage(ActionEnd, uint64(docID), page)}
			forward = &tg.KeyboardButtonCallback{Text: "➡️", Data: preparePage(ActionForward, uint64(docID), page)}
		} else {
			last = &tg.KeyboardButtonCallback{Text: "⏺", Data: []byte{ActionNone}}
			if pages > int(page) {
				forward = &tg.KeyboardButtonCallback{Text: "➡️", Data: preparePage(ActionForward, uint64(docID), page)}
			} else {
				forward = &tg.KeyboardButtonCallback{Text: "⏺", Data: []byte{ActionNone}}
			}
		}

		navButtons := []tg.KeyboardButtonClass{
			first, backward,
			&tg.KeyboardButtonCallback{Text: fmt.Sprintf("%d / %d", page+1, pages+1), Data: []byte{ActionNone}},
			forward, last,
		}

		markup.Rows = make([]tg.KeyboardButtonRow, 2+len(relevantTags))
		markup.Rows[0] = tg.KeyboardButtonRow{Buttons: navButtons}
		markup.Rows[len(relevantTags)+1] = tg.KeyboardButtonRow{Buttons: navButtons}
		for i := 0; i < len(relevantTags); i++ {
			b := prepareTagID(relevantTags[i].ID, uint64(docID), page)
			button := &tg.KeyboardButtonCallback{Text: relevantTags[i].Tag, Data: b}
			markup.Rows[i+1] = tg.KeyboardButtonRow{Buttons: []tg.KeyboardButtonClass{button}}
		}
	}
	return markup, nil
}
