package webapp

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
)

type InitDataField interface {
	Name() string
	EncodeField() ([]byte, error)
}

type InitData []InitDataField

func (hd *InitData) Serialize() ([]byte, error) {
	buf := &bytes.Buffer{}
	for _, f := range *hd {
		if _, err := buf.WriteString(f.Name()); err != nil {
			return nil, err
		}
		if err := buf.WriteByte('='); err != nil {
			return nil, err
		}
		field, err := f.EncodeField()
		if err != nil {
			return nil, err
		}
		if _, err := buf.Write(field); err != nil {
			return nil, err
		}
		if err := buf.WriteByte('&'); err != nil {
			return nil, err
		}
	}
	result := buf.Bytes()
	return result[:buf.Len()-1], nil
}

func (hd *InitData) Sign(key []byte) ([]byte, error) {
	sort.Slice(*hd, func(i, j int) bool {
		return strings.Compare((*hd)[i].Name(), (*hd)[j].Name()) < 0
	})
	b, err := hd.Serialize()
	if err != nil {
		return nil, err
	}

	for i, char := range b {
		if char == '&' {
			b[i] = '\n'
		}
	}
	hm := hmac.New(sha256.New, key)
	if _, err := hm.Write(b); err != nil {
		return nil, err
	}
	for i, char := range b {
		if char == '\n' {
			b[i] = '&'
		}
	}
	hash := Hash(hex.EncodeToString(hm.Sum(nil)))

	buf := bytes.NewBuffer(b)
	if _, err := buf.WriteString("&hash="); err != nil {
		return nil, err
	}
	if _, err := buf.WriteString(string(hash)); err != nil {
		return nil, err
	}
	*hd = append(*hd, &hash)
	return buf.Bytes(), nil
}
