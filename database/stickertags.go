package database

import (
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
// | document_id | bigint       | YES  |     | NULL              |                   |
// | access_hash | bigint       | YES  |     | NULL              |                   |
// | tag         | text         | YES  |     | NULL              |                   |
// | added       | datetime     | NO   |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED |
// +-------------+--------------+------+-----+-------------------+-------------------+

type StickerTag struct {
	ID         uint64 `gorm:"primaryKey"`
	User       int64
	StickerID  uint64
	Sticker    *Sticker
	DocumentID int64
	AccessHash int64
	Tag        string
	Added      time.Time `gorm:"->"` //`sql:"DEFAULT:CURRENT_TIMESTAMP"`
}

func (st *StickerTag) GetAllForUser() (tags []*StickerTag, err error) {
	err = DB.Where(st).Find(&tags).Error
	return tags, err
}

func (st *StickerTag) CheckAndAdd() (response string, err error) {
	var temp StickerTag
	// Access hash can theoretically change, so we can't use it for the search
	err = DB.Where(&StickerTag{User: st.User, DocumentID: st.DocumentID, Tag: st.Tag}).First(&temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return "Что-то пошло не так!", err
	}
	if temp.ID > 0 {
		return "У этого стикера уже есть этот тег, действие отменено", nil
	} else {
		err = DB.Create(st).Error
		if err != nil {
			return "Что-то пошло не так, попробуйте ещё раз!", err
		} else {
			return "Тег добавлен!", nil
		}
	}
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
