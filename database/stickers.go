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
// | type           | tinyint(1)   | NO   |     | NULL              |                   |
// | added          | datetime     | NO   |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED |
// +----------------+--------------+------+-----+-------------------+-------------------+

type MimeType uint8

const (
	MimeTypeWebp = iota
	MimeTypeTgs
	MimeTypeWebm
)

type Sticker struct {
	ID            uint64 `gorm:"primaryKey"`
	DocumentID    int64
	AccessHash    int64
	FileReference []byte
	FileContext   string
	Type          MimeType
	Added         time.Time `gorm:"->"`
}

// Fetch gets or creates sticker record with given DocumentID, AccessHash
// and FileReference
func (s *Sticker) Fetch() error {
	err := DB.Where(&Sticker{DocumentID: s.DocumentID}).
		Attrs(&Sticker{AccessHash: s.AccessHash, FileReference: s.FileReference, Type: s.Type}).
		FirstOrCreate(s).Error
	if err != nil {
		return err
	}
	return nil
}

type ThumbnailSize string

func (t ThumbnailSize) ToString() string {
	return string(t)
}

const (
	Box100px   ThumbnailSize = "s"
	Box320px   ThumbnailSize = "m"
	Box800px   ThumbnailSize = "x"
	Box1280px  ThumbnailSize = "y"
	Box2560px  ThumbnailSize = "w"
	Crop160px  ThumbnailSize = "a"
	Crop320px  ThumbnailSize = "b"
	Crop640px  ThumbnailSize = "c"
	Crop1280px ThumbnailSize = "d"

	Strip   ThumbnailSize = "i"
	Outline ThumbnailSize = "j"

	NoThumbnail = ""
)

// Download downloads sticker file and returns it as bytes
func (s *Sticker) Download(ctx context.Context, api *tg.Client, thumbnailSize ThumbnailSize) ([]byte, error) {
	// we're fine with raw api call since it's not likely we'll get
	// over 512KB in size
	upd, err := api.UploadGetFile(ctx, &tg.UploadGetFileRequest{
		Location: &tg.InputDocumentFileLocation{
			ID:            s.DocumentID,
			AccessHash:    s.AccessHash,
			FileReference: s.FileReference,
			ThumbSize:     thumbnailSize.ToString(),
		},
		Limit: 512 * 1024,
	})
	if err != nil {
		return nil, err
	}
	if uploadFile, ok := upd.(*tg.UploadFile); ok {
		return uploadFile.Bytes, nil
	}
	return nil, err
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
