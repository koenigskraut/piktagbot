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

type InitData []InitDataField

func (hd *InitData) Serialize(glue byte) ([]byte, error) {
	buf := &bytes.Buffer{}
	w := bufio.NewWriter(buf)
	for i, f := range *hd {
		if err := EncodeField(w, f); err != nil {
			return nil, err
		}
		if i == len(*hd)-1 {
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
	if glue == '&' {
		b = []byte(url.PathEscape(string(b)))
	}
	return b, nil
}

func (hd *InitData) Hash(key []byte) (*Hash, error) {
	sort.Slice(*hd, func(i, j int) bool {
		return strings.Compare((*hd)[i].Name(), (*hd)[j].Name()) < 0
	})
	b, err := hd.Serialize('\n')
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

func (hd *InitData) Sign(key []byte) ([]byte, error) {
	hash, err := hd.Hash(key)
	if err != nil {
		return nil, err
	}
	*hd = append(*hd, hash)
	return hd.Serialize('&')
}

func (hd *InitData) Parse(data []byte) error {
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
		*hd = append(*hd, field)
	}
	return nil
}

func (hd *InitData) Verify(key []byte) (bool, error) {
	fieldsNum := len(*hd) - 1
	newData := make(InitData, 0, fieldsNum)
	newData = (*hd)[:fieldsNum]
	hash, err := newData.Hash(key)
	if err != nil {
		return false, err
	}
	return hash.Data == (*hd)[fieldsNum].(*Hash).Data, nil
}

func (hd *InitData) ToMap() map[string]InitDataField {
	m := make(map[string]InitDataField)
	for _, field := range *hd {
		m[field.Name()] = field
	}
	return m
}

// TODO think about it?
//type InitDataStruct struct {
//	QueryID string
//	User    *User
//	Hash    string
//	RawData InitData
//}

//func (hd *InitData) ToStruct() (InitDataStruct, error) {
//	s := InitDataStruct{
//		RawData: make(InitData, 0, 4),
//	}
//	for _, field := range s.RawData {
//		switch v := field.(type) {
//		case *Hash:
//			s.Hash = v.Data
//		case *User:
//			s.User = v
//		}
//	}
//	return s, nil
//}
