package webapp

import (
	"encoding/json"
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

const InitFieldUserName = "user"

func (u *User) Name() string {
	return InitFieldUserName
}

func (u *User) EncodeField() ([]byte, error) {
	return json.Marshal(u)
}

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

func (u *User) DecodeField(input string) error {
	if err := json.Unmarshal([]byte(input), u); err != nil {
		return err
	}
	return nil
}

func ProduceUser(input string) (InitDataField, error) {
	var user User
	if err := user.DecodeField(input); err != nil {
		return nil, err
	}
	return &user, nil
}
