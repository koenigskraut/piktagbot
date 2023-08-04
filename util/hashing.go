package util

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gotd/td/tg"
	"net/http"
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

func HashOfJSON(v any) (string, error) {
	fmt.Printf("hashing %v\n", v)
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	fmt.Printf("marshalled as %s\n", b)
	hm := hmac.New(sha256.New, GetSecretKey())
	hm.Write(b)
	result := hex.EncodeToString(hm.Sum(nil))
	fmt.Printf("hash: %s\n", result)
	return result, nil
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

func (u *WebAppUser) FillFrom(tgUser *tg.User) {
	u.ID = tgUser.ID
	u.FirstName = tgUser.FirstName
	u.LastName = tgUser.LastName
	username, _ := tgUser.GetUsername()
	u.Username = username
	u.LanguageCode = tgUser.LangCode
	u.IsPremium = tgUser.Premium
	u.AllowsWriteToPM = false
}

type WebAppParams struct {
	QueryID  string     `json:"query_id"`
	User     WebAppUser `json:"user"`
	AuthDate string     `json:"auth_date"`
	Hash     string     `json:"hash"`
}

func (wp *WebAppParams) Serialize() (string, error) {
	fields := make([]string, 3)
	user, err := json.Marshal(wp.User)
	if err != nil {
		return "", err
	}
	fields[0] = fmt.Sprintf("query_id=%s", wp.QueryID)
	fields[1] = fmt.Sprintf("user=%s", string(user))
	fields[2] = fmt.Sprintf("auth_date=%s", wp.AuthDate)
	start := strings.Join(fields, "&")

	hash := hashOfFields(fields)
	wp.Hash = hash
	return fmt.Sprintf("%s&hash=%s", url.PathEscape(start), hash), nil
}

func webAppUserFromString(user string) (*WebAppUser, error) {
	buffer := bytes.NewBufferString(user)
	var u WebAppUser
	if err := json.NewDecoder(buffer).Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func ParseInitData(initData []byte) (*WebAppParams, error) {
	values, err := url.ParseQuery(string(initData))
	if err != nil {
		return nil, err
	}
	fields := make([]string, 0, 3) // magic number, it is 3 for now irl
	var params WebAppParams
	for k, v := range values {
		switch k {
		case "hash":
			params.Hash = v[0]
			continue
		case "user":
			err = json.Unmarshal([]byte(v[0]), &params.User)
		case "auth_date":
			params.AuthDate = v[0]
		}
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("%s=%s", k, v[0]))
	}
	sort.Strings(fields)
	toHash := strings.Join(fields, "\n")

	if params.Hash != hashOfString(toHash) {
		return nil, ErrHashMismatch
	}

	return &params, nil
}

func HashOfRequestUser(request *http.Request) string {
	ip := request.RemoteAddr
	userAgent := request.Header.Get("User-Agent")
	toHash := ip[:strings.Index(ip, ":")+1] + userAgent
	return hashOfString(toHash)
}
