package webapp

import (
	"strconv"
)

//type WebAppParams struct {
//	QueryID  string     `json:"query_id"`
//	User     WebAppUser `json:"user"`
//	AuthDate string     `json:"auth_date"`
//	Prefix   string     `json:"prefix"`
//	Hash     string     `json:"hash"`
//}

// QueryID is a unique identifier for the Web App session, required for sending messages via the
// messages.sendWebViewResultMessage¹ method.
//
// Links:
//  1. https://core.telegram.org/method/messages.sendWebViewResultMessage
type QueryID string

func (q QueryID) Name() string {
	return "query_id"
}

func (q QueryID) EncodeField() ([]byte, error) {
	return []byte(q), nil
}

// AuthDate is Unix time when the form was opened.
type AuthDate int64

func (a AuthDate) Name() string {
	return "auth_date"
}

func (a AuthDate) EncodeField() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(a), 10)), nil
}

// Prefix is the user inline query, sent only by this bot from inline mode.
type Prefix string

func (p Prefix) Name() string {
	return "prefix"
}

func (p Prefix) EncodeField() ([]byte, error) {
	return []byte(p), nil
}

// Hash is a hash of all passed parameters, which the bot server can use to check their validity.
type Hash string

func (h Hash) Name() string {
	return "hash"
}

func (h Hash) EncodeField() ([]byte, error) {
	return []byte(h), nil
}
