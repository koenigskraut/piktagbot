package database

import "time"

// TODO rework DB, add file context (although it works as is, it shouldn't) and make use of all of this

type Sticker struct {
	ID            uint64 `gorm:"primaryKey"`
	DocumentID    int64
	AccessHash    int64
	FileReference []byte
	FileContext   string
	Added         time.Time `gorm:"autoCreateTime"`
}

//func (s *Sticker) Download(ctx context.Context, api *tg.Client) ([]byte, error) {
//	upd, err := api.UploadGetFile(ctx, &tg.UploadGetFileRequest{
//		Location: &tg.InputDocumentFileLocation{
//			ID:            s.DocumentID,
//			AccessHash:    s.AccessHash,
//			FileReference: s.FileReference,
//			ThumbSize:     "m",
//		},
//		Limit: 512 * 1024,
//	})
//	return upd.(*tg.UploadFile).Bytes, err
//}
//
//func (s *Sticker) Fetch() (stickerID uint64, err error) {
//	var temp Sticker
//	err = DB.Where(&Sticker{DocumentID: s.DocumentID}).First(&temp).Error
//	if err != nil && err != gorm.ErrRecordNotFound {
//		return 0, err
//	}
//	if temp.ID > 0 {
//		return temp.ID, nil
//	} else {
//		err = DB.Create(s).Error
//		return s.ID, err
//	}
//}
