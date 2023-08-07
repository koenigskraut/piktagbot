package webapp

import "bufio"

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
