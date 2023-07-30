package database

import (
	"context"
	"github.com/gotd/td/tg"
	"time"
)

// TODO rework DB, add file context (although it works as is, it shouldn't) and make use of all of this
// mysql> describe stickers;
// +----------------+--------------+------+-----+-------------------+-------------------+
// | Field          | Type         | Null | Key | Default           | Extra             |
// +----------------+--------------+------+-----+-------------------+-------------------+
// | id             | int unsigned | NO   | PRI | NULL              | auto_increment    |
// | document_id    | bigint       | NO   |     | NULL              |                   |
// | access_hash    | bigint       | NO   |     | NULL              |                   |
// | file_reference | blob         | YES  |     | NULL              |                   |
// | file_context   | text         | YES  |     | NULL              |                   |
// | added          | datetime     | NO   |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED |
// +----------------+--------------+------+-----+-------------------+-------------------+

type Sticker struct {
	ID            uint64 `gorm:"primaryKey"`
	DocumentID    int64
	AccessHash    int64
	FileReference []byte
	FileContext   string
	Added         time.Time `gorm:"->"`
}

// Fetch gets or creates sticker record with given DocumentID, AccessHash
// and FileReference
func (s *Sticker) Fetch() error {
	err := DB.Where(&Sticker{DocumentID: s.DocumentID}).
		Attrs(&Sticker{AccessHash: s.AccessHash, FileReference: s.FileReference}).
		FirstOrCreate(s).Error
	if err != nil {
		return err
	}
	return nil
}

// Download downloads sticker file and returns it as bytes
func (s *Sticker) Download(ctx context.Context, api *tg.Client) ([]byte, error) {
	upd, err := api.UploadGetFile(ctx, &tg.UploadGetFileRequest{
		Location: &tg.InputDocumentFileLocation{
			ID:            s.DocumentID,
			AccessHash:    s.AccessHash,
			FileReference: s.FileReference,
			ThumbSize:     "x",
		},
		Limit: 512 * 1024,
	})
	return upd.(*tg.UploadFile).Bytes, err
}

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
