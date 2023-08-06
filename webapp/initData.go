package webapp

import (
	"bufio"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

type InitDataField interface {
	Name() string
	EncodeData(*bufio.Writer) error
	DecodeData(*bufio.Reader) error
}

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

func (hd *InitData) Sign(key []byte) ([]byte, error) {
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

	buf := bytes.NewBuffer(b)
	w := bufio.NewWriter(buf)
	if err := EncodeField(w, hash); err != nil {
		return nil, err
	}
	if err := w.Flush(); err != nil {
		return nil, err
	}
	*hd = append(*hd, hash)
	return hd.Serialize('&')
}
