package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/koenigskraut/piktagbot/util"
	"io"
	"log"
	"net/http"
)

type SessionUser struct {
	UserID    int64  `json:"user_id"`
	SessionID uint64 `json:"session_id"`
}

func handleVerification(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("hash called")
	b, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "could not read body", http.StatusBadRequest)
		return
	}

	user, err := util.ParseInitData(b)
	if err != nil {
		if errors.Is(err, util.ErrHashMismatch) {
			http.Error(writer, "hash mismatch", http.StatusBadRequest)
		} else {
			http.Error(writer, "malformed init data", http.StatusBadRequest)
		}
		return
	}

	// TODO add sessions
	toReturn, err := json.Marshal(SessionUser{UserID: user.ID, SessionID: 123})
	if err != nil {
		http.Error(writer, "server error", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	_, err = writer.Write(toReturn)
	if err != nil {
		log.Println(err)
		return
	}
}
