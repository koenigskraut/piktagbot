package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

func handleVerification(token string) func(writer http.ResponseWriter, request *http.Request) {

	// required telegram data verification, calculated once based on env variable
	h := hmac.New(sha256.New, []byte("WebAppData"))
	h.Write([]byte(token))
	secretKey := h.Sum(nil)

	return func(writer http.ResponseWriter, request *http.Request) {
		fmt.Println("hash called")
		b, err := io.ReadAll(request.Body)
		if err != nil {
			http.Error(writer, "could not read body", http.StatusBadRequest)
			return
		}

		user, ok, err := processInitData(b, secretKey)
		if err != nil {
			http.Error(writer, "malformed init data", http.StatusBadRequest)
			return
		}
		if !ok {
			http.Error(writer, "hash mismatch", http.StatusBadRequest)
			fmt.Println(3, err)
			return
		}

		_, err = writer.Write([]byte(user))
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func processInitData(data, secretKey []byte) (user string, ok bool, err error) {
	fields := strings.Split(string(data), "&")
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
			user = field[5:]
			fallthrough
		default:
			fieldsToHash = append(fieldsToHash, field)
		}
	}

	sort.Slice(fieldsToHash, func(i, j int) bool {
		return strings.Compare(fieldsToHash[i], fieldsToHash[j]) <= 0
	})

	toHash := strings.Join(fieldsToHash, "\n")

	hm := hmac.New(sha256.New, secretKey)
	hm.Write([]byte(toHash))
	ok = hash == hex.EncodeToString(hm.Sum(nil))

	return
}
