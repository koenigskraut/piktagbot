package webapp

import (
	"bufio"
	"github.com/gotd/td/tg"
)

// User is an object that contains the data of the Web App user.
//
// See https://core.telegram.org/bots/webapps#webappuser for reference.
type User struct {
	ID              int64  `json:"id"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Username        string `json:"username"`
	LanguageCode    string `json:"language_code"`
	IsPremium       bool   `json:"is_premium"`
	AllowsWriteToPM bool   `json:"allows_write_to_pm"`
}

const UserName = "user"

func (u *User) FillFrom(tgUser *tg.User) {
	u.ID = tgUser.ID
	u.FirstName = tgUser.FirstName
	u.LastName = tgUser.LastName
	username, _ := tgUser.GetUsername()
	u.Username = username
	u.LanguageCode = tgUser.LangCode
	u.IsPremium = tgUser.Premium
	u.AllowsWriteToPM = false
}

func (u *User) Name() string {
	return UserName
}

func (u *User) EncodeData(w *bufio.Writer) error {
	return writeJSON(w, u)
}

func (u *User) DecodeData(r *bufio.Reader) error {
	return readJSON(r, u)
}
