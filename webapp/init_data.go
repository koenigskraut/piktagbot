package webapp

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/url"
	"sort"
	"strings"
)

type InitDataMap map[string]InitDataField
type InitDataList []InitDataField

// it is useful to be able to hash and serialize both types, InitDataList is good for initialization and immediate
// signing + serialization, InitDataMap is good for parsing and accessing fields

// Serialize dumps all InitDataList content to string in format field_name=field_data joined by glue
func (dl *InitDataList) Serialize(glue byte) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := bufio.NewWriter(buf)
	for i, f := range *dl {
		if err := EncodeField(w, f); err != nil {
			return nil, err
		}
		if i == len(*dl)-1 {
			continue
		}
		if err := w.WriteByte(glue); err != nil {
			return nil, err
		}
	}
	if err := w.Flush(); err != nil {
		return nil, err
	}
	b := buf.Bytes()
	// glue can be '\n' or '&', '&' is for the URL, so escape path
	if glue == '&' {
		b = []byte(url.PathEscape(string(b)))
	}
	return b, nil
}

// Hash calculates HMAC-SHA256 hash of a given InitData, serialized in a following manner:
//  1. Order fields in alphabetical order by their names
//  2. Stringify individual fields as field_name=field_data
//  3. Join resulting strings with '\n' symbol
func (dl *InitDataList) Hash(key []byte) (*Hash, error) {
	sort.Slice(*dl, func(i, j int) bool {
		return strings.Compare((*dl)[i].Name(), (*dl)[j].Name()) < 0
	})

	b, err := dl.Serialize('\n')
	if err != nil {
		return nil, err
	}

	hm := hmac.New(sha256.New, key)
	if _, err := hm.Write(b); err != nil {
		return nil, err
	}
	hash := &Hash{Data: hex.EncodeToString(hm.Sum(nil))}
	return hash, nil
}

// Hash calculates HMAC-SHA256 hash of a given InitData, serialized in a following manner:
//  1. Order fields in alphabetical order by their names
//  2. Stringify individual fields as field_name=field_data
//  3. Join resulting strings with '\n' symbol
func (dm *InitDataMap) Hash(key []byte) (*Hash, error) {
	dl := dm.ToSlice()
	return dl.Hash(key)
}

// Sign calculates HMAC-SHA256 hash of a given InitDataList as in InitDataList.Hash method and appends it as
// a InitDataField to the slice
func (dl *InitDataList) Sign(key []byte) error {
	hash, err := dl.Hash(key)
	if err != nil {
		return err
	}
	*dl = append(*dl, hash)
	return nil
}

// Parse reads InitDataMap from bytes
func (dm *InitDataMap) Parse(data []byte) error {
	buf := bytes.NewBuffer(data)
	r := bufio.NewReader(buf)
	for {
		field, err := DecodeField(r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		(*dm)[field.Name()] = field
	}
	return nil
}

// Verify recalculates hash for a given InitDataMap and compares it with the existing one
func (dm *InitDataMap) Verify(key []byte) (bool, error) {
	oldHash, ok := (*dm)[HashName].(*Hash)
	if !ok {
		return false, errors.New("init data not signed")
	}
	delete(*dm, HashName)
	newHash, err := dm.Hash(key)
	if err != nil {
		return false, err
	}
	(*dm)[HashName] = oldHash
	return newHash.Data == oldHash.Data, nil
}

// ToSlice returns a copy of InitDataMap as a slice
func (dm *InitDataMap) ToSlice() InitDataList {
	l := make([]InitDataField, 0, len(*dm))
	for _, field := range *dm {
		l = append(l, field)
	}
	return l
}

// ToMap returns a copy of InitDataList as a map
func (dl *InitDataList) ToMap() InitDataMap {
	m := make(map[string]InitDataField)
	for _, field := range *dl {
		m[field.Name()] = field
	}
	return m
}
