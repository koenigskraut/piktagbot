package util

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"sort"
	"strings"
	"sync"
)

var ErrHashMismatch = errors.New("hash mismatch")

var once sync.Once
var secretKey []byte

func GetSecretKey() []byte {
	once.Do(func() {
		// required telegram data verification, calculated once based on env variable
		h := hmac.New(sha256.New, []byte("WebAppData"))
		h.Write([]byte(os.Getenv("BOT_TOKEN")))
		secretKey = h.Sum(nil)
	})
	return secretKey
}

func hashOfString(toHash string) string {
	hm := hmac.New(sha256.New, GetSecretKey())
	hm.Write([]byte(toHash))
	return hex.EncodeToString(hm.Sum(nil))
}

func hashOfFields(fields []string) string {
	sort.Slice(fields, func(i, j int) bool {
		return strings.Compare(fields[i], fields[j]) <= 0
	})
	toHash := strings.Join(fields, "\n")
	return hashOfString(toHash)
}

type WebAppUser struct {
	ID              int64  `json:"id"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Username        string `json:"username"`
	LanguageCode    string `json:"language_code"`
	IsPremium       bool   `json:"is_premium"`
	AllowsWriteToPM bool   `json:"allows_write_to_pm"`
}

func webAppUserFromString(user string) (*WebAppUser, error) {
	buffer := bytes.NewBufferString(user)
	var u WebAppUser
	if err := json.NewDecoder(buffer).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func ParseInitData(initData []byte) (user *WebAppUser, err error) {
	fields := strings.Split(string(initData), "&")
	fieldsToHash := make([]string, 0, len(fields)-1)

	var hash string

	for _, field := range fields {
		switch field[:5] {
		case "hash=":
			hash = field[5:]
		case "user=":
			field, err = url.QueryUnescape(field)
			if err != nil {
				return
			}
			user, err = webAppUserFromString(field[5:])
			if err != nil {
				return
			}
			fallthrough
		default:
			fieldsToHash = append(fieldsToHash, field)
		}
	}

	if hash != hashOfFields(fieldsToHash) {
		return nil, ErrHashMismatch
	}

	return
}