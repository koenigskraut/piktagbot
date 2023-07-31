package util

import (
	"github.com/gotd/td/tg"
	"github.com/koenigskraut/piktagbot/database"
)

// StickerFromMedia safely unpacks tg.MessageMediaClass, returns true
// if it is a sticker (webp/tgs/webm) and writes it into the database
func StickerFromMedia(media tg.MessageMediaClass) (*database.Sticker, bool) {
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

	if !isSticker {
		return nil, false
	}

	var s *database.Sticker
	s = &database.Sticker{
		DocumentID:    document.ID,
		AccessHash:    document.AccessHash,
		FileReference: document.FileReference,
	}
	switch document.MimeType {
	case "image/webp":
		s.Type = database.MimeTypeWebp
	case "application/x-tgsticker":
		s.Type = database.MimeTypeTgs
	case "video/webm":
		s.Type = database.MimeTypeWebm
	default:
		return nil, false
	}
	_ = s.Fetch()

	return s, true
}
