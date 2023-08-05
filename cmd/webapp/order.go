package main

import (
	"encoding/json"
	"fmt"
	db "github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/util"
	"net/http"
)

type TagOrderUpdate struct {
	UserSession Session `json:"session"`
	Order       []int64 `json:"order"`
	NewFirst    bool    `json:"new_first"`
}

func handleOrderUpdate(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("order called")
	var update TagOrderUpdate
	dec := json.NewDecoder(request.Body)
	if dec.Decode(&update) != nil {
		http.Error(writer, "malformed json", http.StatusBadRequest)
		return
	}
	// verification
	hashCompareTo := update.UserSession.Hash
	update.UserSession.Hash = ""
	hashCalculated, err := util.HashOfJSON(update.UserSession)
	if err != nil {
		http.Error(writer, "server error", http.StatusInternalServerError)
		return
	}
	if hashCalculated != hashCompareTo {
		http.Error(writer, "hash mismatch", http.StatusBadRequest)
		return
	}
	// get session
	verifiedSession, ok := oneTimeSessions.peek(update.UserSession.SessionID)
	if !ok {
		http.Error(writer, "session expired", http.StatusBadRequest)
		return
	}
	// check ip and user-agent
	userHash := util.HashOfRequestUser(request)
	if verifiedSession.clientHash != userHash {
		http.Error(writer, "bad user", http.StatusBadRequest)
		return
	}
	// get user
	dbUser := db.User{UserID: verifiedSession.userID}
	if _, err := dbUser.Get(); err != nil {
		http.Error(writer, "server error", http.StatusInternalServerError)
		return
	}

	order, err := dbUser.GetOrder("")
	if err != nil {
		http.Error(writer, "server error", http.StatusInternalServerError)
		return
	}
	b, err := db.WriteTagOrder(update.Order)
	if err != nil {
		http.Error(writer, "server error", http.StatusInternalServerError)
		return
	}
	order.TagOrder, order.NewFirst = b, update.NewFirst
	if err := order.Save(); err != nil {
		http.Error(writer, "server error", http.StatusInternalServerError)
		return
	}
}
