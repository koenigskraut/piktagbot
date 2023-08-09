package database

import (
	"gorm.io/gorm"
	"time"
)

const GlobalTagDefault = false

// mysql> describe users;
// +------------+--------------+------+-----+-------------------+-------------------+
// | Field      | Type         | Null | Key | Default           | Extra             |
// +------------+--------------+------+-----+-------------------+-------------------+
// | id         | int unsigned | NO   | PRI | NULL              | auto_increment    |
// | user_id    | bigint       | NO   |     | NULL              |                   |
// | global_tag | tinyint(1)   | YES  |     | NULL              |                   |
// | flag       | text         | YES  |     | NULL              |                   |
// | flag_data  | text         | YES  |     | NULL              |                   |
// | tag_order  | blob         | YES  |     | NULL              |                   |
// | new        | tinyint(1)   | NO   |     | NULL              |                   |
// | added      | datetime     | YES  |     | CURRENT_TIMESTAMP | DEFAULT_GENERATED |
// +------------+--------------+------+-----+-------------------+-------------------+

type User struct {
	ID        uint `gorm:"primaryKey"`
	UserID    int64
	GlobalTag bool
	Flag      int8
	FlagData  string
	New       bool
	TagOrder  []byte
	Added     time.Time `gorm:"->"`
}

// Get fetches a user record from DB or creates if it doesn't exist
func (u *User) Get() (err error) {
	if err = DB.Where(&User{UserID: u.UserID}).First(u).Error; err == nil {
		return
	} else if err == gorm.ErrRecordNotFound {
		*u = User{UserID: u.UserID, GlobalTag: GlobalTagDefault, New: true}
		err = DB.Create(u).Error
	}
	return
}

func (u *User) Save() error {
	return DB.Save(u).Error
}

func (u *User) SwitchGlobal() (err error) {
	return DB.Model(u).Update("global_tag", !u.GlobalTag).Error
}

func (u *User) SetFlag(flag int8, flagData string) (err error) {
	err = DB.Model(u).Updates(&User{Flag: flag, FlagData: flagData}).Error
	return
}

func (u *User) RecentStickers() ([]*StickerTag, error) {
	var pre []*StickerTag
	err := DB.Model(&StickerTag{}).
		Preload("Sticker").
		Select("sticker_id, MAX(added) added").
		Where(&StickerTag{User: u.UserID}).
		Order("added desc").
		Group("sticker_id").
		Limit(500).
		Find(&pre).Error
	if err != nil {
		return nil, err
	}
	order, err := u.GetOrder("")
	if err != nil {
		return nil, err
	}
	err = stickersSort(pre, order)
	if err != nil {
		return nil, err
	}
	return pre, nil
}

func (u *User) SearchStickers(prefix string) ([]*StickerTag, error) {
	query := DB.Preload("Sticker").
		Select("sticker_id, MAX(added) added").
		Order("added desc").
		Where("tag LIKE ?", prefix+"%").
		Order("added desc").
		Group("sticker_id").
		Limit(500)
	if u.GlobalTag {
		query = query.Where("user IN (?, 0)", u.UserID)
	} else {
		query = query.Where(&StickerTag{User: u.UserID})
	}
	var pre []*StickerTag
	if err := query.Find(&pre).Error; err != nil {
		return nil, err
	}
	order, err := u.GetOrder(prefix)
	if err != nil {
		return nil, err
	}
	err = stickersSort(pre, order)
	if err != nil {
		return nil, err
	}
	return pre, nil
}

func (u *User) GetOrder(prefix string) (*StickerOrder, error) {
	var order StickerOrder
	err := DB.Where(&StickerOrder{
		User:   u.UserID,
		Prefix: prefix,
	}).First(&order).Error
	if err == nil {
		return &order, nil
	}
	if err == gorm.ErrRecordNotFound {
		order = StickerOrder{
			User:     u.UserID,
			Prefix:   prefix,
			NewFirst: true,
		}
		if err = DB.Create(&order).Error; err != nil {
			return nil, err
		}
		return &order, nil
	}
	return nil, err
}

func (u *User) GetNewFirst(prefix string) (bool, error) {
	order, err := u.GetOrder(prefix)
	if err != nil {
		return false, err
	}
	return order.NewFirst, err
}
