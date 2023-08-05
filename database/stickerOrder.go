package database

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// mysql> describe sticker_orders;
// +-----------+--------------+------+-----+---------+----------------+
// | Field     | Type         | Null | Key | Default | Extra          |
// +-----------+--------------+------+-----+---------+----------------+
// | id        | int unsigned | NO   | PRI | NULL    | auto_increment |
// | user      | bigint       | NO   |     | NULL    |                |
// | prefix    | text         | YES  |     | NULL    |                |
// | new_first | tinyint(1)   | NO   |     | 1       |                |
// | tag_order | blob         | YES  |     | NULL    |                |
// +-----------+--------------+------+-----+---------+----------------+

type StickerOrder struct {
	ID       uint64 `gorm:"primaryKey"`
	User     int64
	Prefix   string
	NewFirst bool
	TagOrder []byte
}

func readTagOrder(serialized []byte) (tagOrder []int64, err error) {
	tagOrder = make([]int64, len(serialized)/8)
	buf := bytes.NewReader(serialized)
	err = binary.Read(buf, binary.LittleEndian, tagOrder)
	return
}

func writeTagOrder(order []int64) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, order)
	return buf.Bytes(), err
}

func ReadTagOrder(serialized []byte) (tagOrder []int64, err error) {
	return readTagOrder(serialized)
}

func WriteTagOrder(order []int64) (serialized []byte, err error) {
	return writeTagOrder(order)
}

func sortByOrder(toSort []*StickerTag, order []int64, unorderedFirst bool) {
	sortedMap := make(map[int64]*StickerTag)
	for _, s := range toSort {
		sortedMap[s.Sticker.DocumentID] = s
	}
	results := make([]*StickerTag, 0, len(toSort))
	for _, o := range order {
		if s, ok := sortedMap[o]; ok {
			results = append(results, s)
			delete(sortedMap, o)
		}
	}
	i := 0
	for _, v := range toSort {
		if s, ok := sortedMap[v.Sticker.DocumentID]; ok {
			toSort[i] = s
			i++
		}
	}
	if unorderedFirst {
		copy(toSort[i:], results)
	} else {
		copy(toSort[len(toSort)-i:], toSort[:i])
		copy(toSort, results)
	}
}

func (so *StickerOrder) SortStickers(stickers []*StickerTag) error {
	order, err := readTagOrder(so.TagOrder)
	if err != nil {
		fmt.Println(err)
		return err
	}
	sortByOrder(stickers, order, so.NewFirst)
	return nil
}

func (so *StickerOrder) Save() error {
	return DB.Save(so).Error
}

func (so *StickerOrder) UpdateFromOrder(order []int64) error {
	serialized, err := writeTagOrder(order)
	if err != nil {
		return err
	}
	return DB.Model(so).Where(&StickerOrder{ID: so.ID}).Update("tag_order", serialized).Error
}

func (so *StickerOrder) UpdateFromStickers(stickers []*StickerTag) error {
	order := make([]int64, len(stickers))
	for i, s := range stickers {
		order[i] = s.Sticker.DocumentID
	}
	return so.UpdateFromOrder(order)
}
