package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/koenigskraut/piktagbot/util"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Session struct {
	SessionID string `json:"session_id"`
	Prefix    string `json:"prefix"`
	Hash      string `json:"hash,omitempty"`
}

func handleVerification(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("hash called")
	b, err := io.ReadAll(request.Body)
	if err != nil {
		http.Error(writer, "could not read body", http.StatusBadRequest)
		return
	}

	params, err := util.ParseInitData(b)
	if err != nil {
		if errors.Is(err, util.ErrHashMismatch) {
			http.Error(writer, "hash mismatch", http.StatusBadRequest)
		} else {
			http.Error(writer, "malformed init data", http.StatusBadRequest)
		}
		return
	}
	authDate, err := strconv.ParseInt(params.AuthDate, 10, 64)
	userHash := util.HashOfRequestUser(request)
	if delta := time.Now().Unix() - authDate; delta > 30 {
		// session init is definitely failed, maybe user just refreshed the page?
		session, sessionExists := oneTimeSessions.peek(params.Hash[:16])
		// session expired, abort
		if delta > sessionExpirationLimit.Milliseconds()/1000 || !sessionExists {
			http.Error(writer, "date too old", http.StatusBadRequest)
			return
		}
		// client changed user-agent, abort
		if session.clientHash != userHash {
			http.Error(writer, "user-agent mismatch", http.StatusBadRequest)
			return
		}
	}

	oneTimeSessions.createSession(params.User.ID, params.Hash[:16], userHash)
	sessionJSON := Session{SessionID: params.Hash[:16], Prefix: params.Prefix}
	sessionHash, err := util.HashOfJSON(sessionJSON)
	if err != nil {
		http.Error(writer, "server error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	sessionJSON.Hash = sessionHash
	toReturn, err := json.Marshal(sessionJSON)
	if err != nil {
		http.Error(writer, "server error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	fmt.Println("sending json: " + string(toReturn))

	_, err = writer.Write(toReturn)
	if err != nil {
		log.Println(err)
		return
	}
}
