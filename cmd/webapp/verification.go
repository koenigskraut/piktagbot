package main

import (
	"encoding/json"
	"fmt"
	"github.com/koenigskraut/piktagbot/util"
	"github.com/koenigskraut/piktagbot/webapp"
	"io"
	"log"
	"net/http"
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

	params := webapp.InitData{}
	if err := params.Parse(b); err != nil {
		http.Error(writer, "malformed init data", http.StatusBadRequest)
		return
	}
	verified, err := params.Verify(util.GetSecretKey())
	if err != nil {
		http.Error(writer, "malformed init data", http.StatusBadRequest)
		return
	}
	if !verified {
		http.Error(writer, "hash mismatch", http.StatusBadRequest)
	}

	resMap := params.ToMap()
	authDate := resMap[webapp.AuthDateName].(*webapp.AuthDate).Data
	hash := resMap[webapp.HashName].(*webapp.Hash).Data
	user := resMap[webapp.UserName].(*webapp.User)
	query := resMap[webapp.PrefixName].(*webapp.Prefix).Data
	userHash := util.HashOfRequestUser(request)
	if delta := time.Now().Unix() - authDate; delta > 30 {
		// session init is definitely failed, maybe user just refreshed the page?
		session, sessionExists := oneTimeSessions.peek(hash[:16])
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

	oneTimeSessions.createSession(user.ID, hash[:16], userHash)
	sessionJSON := Session{SessionID: hash[:16], Prefix: query}
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
