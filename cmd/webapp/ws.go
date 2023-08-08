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
		if err := dbUser.Get(); err != nil {
			log.Println(err)
			return
		}
		searchPrefix := sessionJSON.Prefix
		var stickers []*db.StickerTag
		if searchPrefix == "" {
			stickers, err = dbUser.RecentStickers()
		} else {
			stickers, err = dbUser.SearchStickers(searchPrefix)
		}
		if err != nil {
			log.Println(err)
			return
		}
		locations := make([]InputDocumentMimeTyped, 0, len(stickers))
		for _, r := range stickers {
			locations = append(locations, InputDocumentMimeTyped{
				mimeType: r.Sticker.Type,
				doc: &tg.InputDocumentFileLocation{
					ID:            r.Sticker.DocumentID,
					AccessHash:    r.Sticker.AccessHash,
					FileReference: r.Sticker.FileReference,
				},
			})
		}
		var newFirst []byte
		newFirstBool, err := dbUser.GetNewFirst(searchPrefix)
		if err != nil {
			log.Println(err)
			return
		}
		if newFirstBool {
			newFirst = []byte("nF: 1")
		} else {
			newFirst = []byte("nF: 0")
		}

		writer := wsutil.NewWriterSize(conn, ws.StateServerSide, ws.OpBinary, 4096)
		myChan := make(chan error)
		myFiles := receiveFiles{
			files:  locations,
			ch:     myChan,
			output: writer,
		}
		if err := wsutil.WriteServerBinary(conn, newFirst); err != nil {
			log.Println(err)
			return
		}
		downloadChan <- &myFiles
		if err1 := <-myChan; err1 != nil {
			log.Println(err1)
			if err2 := wsutil.WriteServerText(conn, []byte(fmt.Sprintf("error: %+v", err1))); err2 != nil {
				log.Println(err2)
			}
			return
		}
		writer.Flush()
	}()
}
