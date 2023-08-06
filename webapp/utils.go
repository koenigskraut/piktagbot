package webapp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
)

func readBytes(r *bufio.Reader, until byte) ([]byte, error) {
	b, err := r.ReadSlice(until)
	if err == nil {
		b = b[:len(b)-1]
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	return b, nil
}

func writeBytes(w *bufio.Writer, b []byte) error {
	_, err := w.Write(b)
	return err
}

func readString(r *bufio.Reader) (string, error) {
	b, err := readBytes(r, '&')
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func writeString(w *bufio.Writer, s string) error {
	_, err := w.WriteString(s)
	return err
}

func readInt(r *bufio.Reader, base int, bitSize int) (int64, error) {
	b, err := readString(r)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(b, base, bitSize)
}

func writeInt(w *bufio.Writer, n int64) error {
	_, err := fmt.Fprintf(w, "%d", n)
	return err
}

func readJSON(r *bufio.Reader, to any) error {
	b, err := readBytes(r, '&')
	if err != nil {
		return err
	}
	unescaped, _ := url.PathUnescape(string(b))
	return json.Unmarshal([]byte(unescaped), to)
}

func writeJSON(w *bufio.Writer, from any) error {
	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(from); err != nil {
		return err
	}
	buf.Truncate(buf.Len() - 1)
	_, err := io.Copy(w, &buf)
	return err
}

func readName(r *bufio.Reader) (string, error) {
	b, err := readBytes(r, '=')
	if err != nil {
		return "", err
	}
	if len(b) < 1 {
		return "", io.EOF
	}
	return string(b), err
}
