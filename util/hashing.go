package util

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
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

func HashOfJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	hm := hmac.New(sha256.New, GetSecretKey())
	hm.Write(b)
	result := hex.EncodeToString(hm.Sum(nil))
	return result, nil
}

func HashOfRequestUser(request *http.Request) string {
	ip := request.RemoteAddr
	userAgent := request.Header.Get("User-Agent")
	toHash := ip[:strings.Index(ip, ":")+1] + userAgent
	return hashOfString(toHash)
}
