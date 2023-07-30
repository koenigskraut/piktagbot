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

func (u *User) RecentStickers() (found []*StickerTag, err error) {
	err = DB.Preload("Sticker").
		Where(&StickerTag{User: u.UserID}).
		Order("added desc").
		Find(&found).Error
	if err != nil {
		return nil, err
	}

	// TODO order stuff
	//var order []uint64
	//order, err = u.ReadTagOrder()
	//if err == nil && len(order) > 0 {
	//	mapped := make(map[uint64]*StickerTag)
	//	for _, f := range found {
	//		mapped[f.ID] = f
	//	}
	//	sorted := make([]*StickerTag, 0, len(found))
	//	for _, o := range order {
	//		sorted = append(sorted, mapped[o])
	//		delete(mapped, o)
	//	}
	//	for _, v := range mapped {
	//		sorted = append(sorted, v)
	//	}
	//	found = sorted
	//}
	//
	//stSet := make(map[uint64]struct{})
	//for i := 0; i < len(found); {
	//	if _, ok := stSet[found[i].StickerID]; !ok {
	//		stSet[found[i].StickerID] = struct{}{}
	//		i++
	//	} else {
	//		found = append(found[:i], found[i+1:]...)
	//	}
	//}

	return
}

func (u *User) SearchStickers(prefix string) (found []*StickerTag, err error) {
	query := DB.Preload("Sticker").
		Order("added desc").
		Where("tag LIKE ?", prefix+"%")
	if u.GlobalTag {
		query = query.Where("user IN (?, 0)", u.UserID)
	} else {
		query = query.Where(&StickerTag{User: u.UserID})
	}
	err = query.Find(&found).Error
	if err != nil {
		return nil, err
	}
	return
}
