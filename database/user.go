package database

import (
	"gorm.io/gorm"
)

const GlobalTagDefault = false

// mysql> describe users;
// +------------+--------------+------+-----+---------+----------------+
// | Field      | Type         | Null | Key | Default | Extra          |
// +------------+--------------+------+-----+---------+----------------+
// | id         | int unsigned | NO   | PRI | NULL    | auto_increment |
// | user_id    | bigint       | NO   |     | NULL    |                |
// | global_tag | tinyint(1)   | YES  |     | NULL    |                |
// | flag       | text         | YES  |     | NULL    |                |
// | flag_data  | text         | YES  |     | NULL    |                |
// | tag_order  | blob         | YES  |     | NULL    |                |
// +------------+--------------+------+-----+---------+----------------+

type User struct {
	ID        uint `gorm:"primaryKey"`
	UserID    int64
	GlobalTag bool
	Flag      string
	FlagData  string
	TagOrder  []byte
}

// Get fetches a user record from DB or creates if it doesn't exist
func (u *User) Get() (new bool, err error) {
	if err = DB.Where(&User{UserID: u.UserID}).First(u).Error; err == nil {
		return
	} else if err == gorm.ErrRecordNotFound {
		new = true
		*u = User{UserID: u.UserID, GlobalTag: GlobalTagDefault}
		err = DB.Create(u).Error
	}
	return
}

func (u *User) SwitchGlobal() (err error) {
	return DB.Model(u).Update("global_tag", !u.GlobalTag).Error
}

func (u *User) SetFlag(flag, flagData string) (err error) {
	err = DB.Model(u).Updates(&User{Flag: flag, FlagData: flagData}).Error
	return
}

// TODO order stuff
//func (u *User) WriteTagOrder(tagOrder []uint64) (err error) {
//	buf := &bytes.Buffer{}
//	err = binary.Write(buf, binary.LittleEndian, tagOrder)
//	if err != nil {
//		return
//	}
//	u.TagOrder = buf.Bytes()
//	err = DB.Save(u).Error
//	return
//}

// TODO order stuff
//func (u *User) ReadTagOrder() (tagOrder []uint64, err error) {
//	tagOrder = make([]uint64, len(u.TagOrder)/8)
//	buf := bytes.NewReader(u.TagOrder)
//	err = binary.Read(buf, binary.LittleEndian, tagOrder)
//	return
//}

func (u *User) RecentStickers() ([]*StickerTag, error) {
	var pre []*StickerTag
	err := DB.Preload("Sticker").
		Where(&StickerTag{User: u.UserID}).
		Order("added desc").
		Find(&pre).Error
	if err != nil {
		return nil, err
	}
	unique := StickerTagsUnique(pre)
	order, err := u.GetOrder("")
	if err != nil {
		return nil, err
	}
	if err := order.SortStickers(unique); err != nil {
		return nil, err
	}
	if err := order.UpdateFromStickers(unique); err != nil {
		return nil, err
	}
	return unique, nil
}

func (u *User) SearchStickers(prefix string) ([]*StickerTag, error) {
	query := DB.Preload("Sticker").
		Order("added desc").
		Where("tag LIKE ?", prefix+"%")
	if u.GlobalTag {
		query = query.Where("user IN (?, 0)", u.UserID)
	} else {
		query = query.Where(&StickerTag{User: u.UserID})
	}
	var pre []*StickerTag
	if err := query.Find(&pre).Error; err != nil {
		return nil, err
	}

	return StickerTagsUnique(pre), nil
}

func (u *User) GetOrder(tag string) (*StickerOrder, error) {
	var order StickerOrder
	err := DB.Where(&StickerOrder{
		User: u.UserID,
		Tag:  tag,
	}).First(&order).Error
	if err == nil {
		return &order, nil
	}
	if err == gorm.ErrRecordNotFound {
		order = StickerOrder{
			User:     u.UserID,
			Tag:      tag,
			NewFirst: true,
		}
		if err := DB.Create(&order).Error; err != nil {
			return nil, err
		}
		var stickerTags []*StickerTag
		if tag == "" {
			stickerTags, err = u.RecentStickers()
			if err != nil {
				return nil, err
			}
		} else {
			stickerTags, err = u.SearchStickers(tag)
			if err != nil {
				return nil, err
			}
		}
		if err = order.UpdateFromStickers(stickerTags); err != nil {
			return nil, err
		}
	}
	return nil, err
}

func (u *User) GetNewFirst(tag string) (bool, error) {
	order, err := u.GetOrder(tag)
	if err != nil {
		return false, err
	}
	return order.NewFirst, err
}
