package util

import (
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/database"
)

// DocFromMedia safely unpacks tg.MessageMediaClass and additionally
// returns true if it is a sticker (webp/tgs/webm)
func DocFromMedia(media tg.MessageMediaClass) (*tg.Document, bool) {
	if media == nil || media.TypeID() != tg.MessageMediaDocumentTypeID {
		return nil, false
	}
	document, ok := media.(*tg.MessageMediaDocument).Document.AsNotEmpty()
	if !ok {
		return nil, false
	}
	isSticker := false
	for _, attribute := range document.Attributes {
		if attribute.TypeID() == tg.DocumentAttributeStickerTypeID {
			isSticker = true
			break
		}
	}
	// add sticker to database
	if isSticker {
		st := database.Sticker{
			DocumentID:    document.ID,
			AccessHash:    document.AccessHash,
			FileReference: document.FileReference,
		}
		_ = st.Fetch()

	}
	return document, isSticker
}
