package database

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

// mysql> describe sticker_tags;
// +-------------+--------------+------+-----+-------------------+-------------------+
// | Field       | Type         | Null | Key | Default           | Extra             |
// +-------------+--------------+------+-----+-------------------+-------------------+
// | id          | int unsigned | NO   | PRI | NULL              | auto_increment    |
// | user        | bigint       | NO   |     | NULL              |                   |
// | sticker_id  | bigint       | NO   |     | NULL              |                   |
// | tag         | text         | YES  |     | NULL              |                   |
// | added       | datetime     | NO   |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED |
// +-------------+--------------+------+-----+-------------------+-------------------+

type StickerTag struct {
	ID        uint64 `gorm:"primaryKey"`
	User      int64
	StickerID uint64
	Sticker   *Sticker
	Tag       string
	Added     time.Time `gorm:"->"` //`sql:"DEFAULT:CURRENT_TIMESTAMP"`
}

func (st *StickerTag) GetAllForUser() (tags []*StickerTag, err error) {
	err = DB.Where(st).Find(&tags).Error
	return tags, err
}

var StickerTagExists = errors.New("sticker-tag pair exists")

func (st *StickerTag) CheckAndAdd() error {
	var temp StickerTag
	err := DB.Where(&StickerTag{User: st.User, StickerID: st.StickerID, Tag: st.Tag}).First(&temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if temp.ID > 0 {
		return StickerTagExists
	} else {
		err = DB.Create(st).Error
		if err != nil {
			return err
		} else {
			return nil
		}
	}
}

func (st *StickerTag) CheckForSet() (bool, error) {
	var temp StickerTag
	err := DB.Preload("Sticker").
		Where(&StickerTag{
			User:    st.User,
			Sticker: &Sticker{StickerSet: st.Sticker.StickerSet},
			Tag:     st.Tag,
			AsSet:   true,
		}).First(&temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}
	return temp.ID > 0, nil
}

// what is this?
//func (st *StickerTagNew) Read() {
//
//	var temp Sticker
//	err = DB.Where(&Sticker{DocumentID: s.DocumentID, AccessHash: s.AccessHash}).First(&temp).Error
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
