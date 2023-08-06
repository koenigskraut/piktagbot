package webapp

import (
	"bufio"
	"errors"
	"fmt"
	"sync"
)

// QueryID is a unique identifier for the Web App session, required for sending messages via the
// messages.sendWebViewResultMessage¹ method.
//
// Links:
//  1. https://core.telegram.org/method/messages.sendWebViewResultMessage
type QueryID struct {
	Data string
}

const QueryIDName = "query_id"

func (q *QueryID) Name() string {
	return QueryIDName
}

func (q *QueryID) EncodeData(w *bufio.Writer) error {
	return writeString(w, q.Data)
}

func (q *QueryID) DecodeData(r *bufio.Reader) error {
	s, err := readString(r)
	if err != nil {
		return err
	}
	q.Data = s
	return nil
}

// AuthDate is Unix time when the form was opened.
type AuthDate struct {
	Data int64
}

const AuthDateName = "auth_date"

func (a *AuthDate) Name() string {
	return AuthDateName
}

func (a *AuthDate) EncodeData(w *bufio.Writer) error {
	return writeInt(w, a.Data)
}

func (a *AuthDate) DecodeData(r *bufio.Reader) error {
	n, err := readInt(r, 10, 64)
	if err != nil {
		return err
	}
	a.Data = n
	return nil
}

// Prefix is the user inline query, sent only by this bot from inline mode.
type Prefix struct {
	Data string
}

const PrefixName = "prefix"

func (p *Prefix) Name() string {
	return PrefixName
}

func (p *Prefix) EncodeData(w *bufio.Writer) error {
	return writeString(w, p.Data)
}

func (p *Prefix) DecodeData(r *bufio.Reader) error {
	s, err := readString(r)
	if err != nil {
		return err
	}
	p.Data = s
	return nil
}

// Hash is a hash of all passed parameters, which the bot server can use to check their validity.
type Hash struct {
	Data string
}

const HashName = "hash"

func (h *Hash) Name() string {
	return HashName
}

func (h *Hash) EncodeData(w *bufio.Writer) error {
	return writeString(w, h.Data)
}

func (h *Hash) DecodeData(r *bufio.Reader) error {
	s, err := readString(r)
	if err != nil {
		return err
	}
	h.Data = s
	return nil
}

func EncodeField(w *bufio.Writer, field InitDataField) error {
	_, err := fmt.Fprintf(w, "%s=", field.Name())
	if err != nil {
		return err
	}
	return field.EncodeData(w)
}

func DecodeField(r *bufio.Reader) (InitDataField, error) {
	name, err := readName(r)
	if err != nil {
		return nil, err
	}
	fieldFunc, ok := getMapTypes()[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("unknown field name: \"%s\"", name))
	}
	field := fieldFunc()
	if err := field.DecodeData(r); err != nil {
		return nil, err
	}
	return field, nil
}

var mapTypes map[string]func() InitDataField
var mapTypesOnce = sync.Once{}

func getMapTypes() map[string]func() InitDataField {
	mapTypesOnce.Do(func() {
		mapTypes = map[string]func() InitDataField{
			QueryIDName:  func() InitDataField { return &QueryID{} },
			AuthDateName: func() InitDataField { return &AuthDate{} },
			PrefixName:   func() InitDataField { return &Prefix{} },
			UserName:     func() InitDataField { return &User{} },
			HashName:     func() InitDataField { return &Hash{} },
		}
	})
	return mapTypes
}
