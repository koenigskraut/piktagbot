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
type QueryID struct {
	Data string
}

const InitFieldQueryIDName = "query_id"

func (q *QueryID) Name() string {
	return InitFieldQueryIDName
}

func (q *QueryID) EncodeField() ([]byte, error) {
	return []byte(q.Data), nil
}

func (q *QueryID) DecodeField(input string) error {
	q.Data = input
	return nil
}

func ProduceQueryID(input string) (InitDataField, error) {
	var queryID QueryID
	if err := queryID.DecodeField(input); err != nil {
		return nil, err
	}
	return &queryID, nil
}

// AuthDate is Unix time when the form was opened.
type AuthDate struct {
	Data int64
}

const InitFieldAuthDateName = "auth_date"

func (a *AuthDate) Name() string {
	return InitFieldAuthDateName
}

func (a *AuthDate) EncodeField() ([]byte, error) {
	return []byte(strconv.FormatInt(a.Data, 10)), nil
}

func (a *AuthDate) DecodeField(input string) error {
	n, err := strconv.ParseInt(input, 10, 64)
	a.Data = n
	return err
}

func ProduceAuthDate(input string) (InitDataField, error) {
	var authDate AuthDate
	if err := authDate.DecodeField(input); err != nil {
		return nil, err
	}
	return &authDate, nil
}

// Prefix is the user inline query, sent only by this bot from inline mode.
type Prefix struct {
	Data string
}

const InitFieldPrefixName = "prefix"

func (p *Prefix) Name() string {
	return InitFieldPrefixName
}

func (p *Prefix) EncodeField() ([]byte, error) {
	return []byte(p.Data), nil
}

func (p *Prefix) DecodeField(input string) error {
	p.Data = input
	return nil
}

func ProducePrefix(input string) (InitDataField, error) {
	var prefix Prefix
	if err := prefix.DecodeField(input); err != nil {
		return nil, err
	}
	return &prefix, nil
}

// Hash is a hash of all passed parameters, which the bot server can use to check their validity.
type Hash struct {
	Data string
}

const InitFieldHashName = "hash"

func (h *Hash) Name() string {
	return InitFieldHashName
}

func (h *Hash) EncodeField() ([]byte, error) {
	return []byte(h.Data), nil
}

func (h *Hash) DecodeField(input string) error {
	h.Data = input
	return nil
}

func ProduceHash(input string) (InitDataField, error) {
	var hash Hash
	if err := hash.DecodeField(input); err != nil {
		return nil, err
	}
	return &hash, nil
}
