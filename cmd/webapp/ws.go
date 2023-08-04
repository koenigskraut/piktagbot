package main

import (
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gotd/td/tg"
	db "github.com/koenigskraut/piktagbot/database"
	"github.com/koenigskraut/piktagbot/util"
	"log"
	"net/http"
)

func handleWS(writer http.ResponseWriter, request *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(request, writer)
	if err != nil {
		log.Println(1, err)
		return
	}
	fmt.Println("ws called")
	go func() {
		defer conn.Close()
		data, err := wsutil.ReadClientText(conn)
		if err != nil {
			log.Println(2, err)
			return
		}
		fmt.Printf("FROM CLIENT: %s\n", data)

		var sessionJSON Session
		if err := json.Unmarshal(data, &sessionJSON); err != nil {
			log.Println(err)
			return
		}
		// verification
		hashCompareTo := sessionJSON.Hash
		sessionJSON.Hash = ""
		hashCalculated, err := util.HashOfJSON(sessionJSON)
		if err != nil {
			log.Println(err)
			return
		}
		if hashCalculated != hashCompareTo {
			log.Println("hash mismatch")
			return
		}
		// get session
		session, ok := oneTimeSessions.peek(sessionJSON.SessionID)
		if !ok {
			log.Println("no session")
			return
		}
		// check ip and user-agent
		userHash := util.HashOfRequestUser(request)
		if session.clientHash != userHash {
			log.Println("bad user")
			return
		}
		// get user
		dbUser := db.User{UserID: session.userID}
		if _, err := dbUser.Get(); err != nil {
			log.Println(err)
			return
		}
		recentStickers, _ := dbUser.RecentStickers()
		locations := make([]InputDocumentMimeTyped, 0, len(recentStickers))
		for _, r := range recentStickers {
			locations = append(locations, InputDocumentMimeTyped{
				mimeType: r.Sticker.Type,
				doc: &tg.InputDocumentFileLocation{
					ID:            r.Sticker.DocumentID,
					AccessHash:    r.Sticker.AccessHash,
					FileReference: r.Sticker.FileReference,
				},
			})
		}
		fmt.Println(locations)

		myChan := make(chan string)
		myFiles := receiveFiles{
			files: locations,
			ch:    myChan,
		}
		downloadChan <- &myFiles
		fmt.Println("sent")
		for r := range myChan {
			fmt.Println(r)
			if err := wsutil.WriteServerText(conn, []byte(r)); err != nil {
				log.Println(3, err)
			}
		}
	}()
}
