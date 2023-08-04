package main

import (
	"encoding/json"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/gotd/td/tg"
	db "github.com/koenigskraut/piktagbot/database"
	"log"
	"net/http"
)

func handleWS(writer http.ResponseWriter, request *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(request, writer)
	if err != nil {
		log.Println(1, err)
		return
	}
	go func() {
		defer conn.Close()
		data, err := wsutil.ReadClientText(conn)
		if err != nil {
			log.Println(2, err)
			return
		}
		fmt.Printf("FROM CLIENT: %s\n", data)

		var sessionUser SessionUser
		if err := json.Unmarshal(data, &sessionUser); err != nil {
			log.Println(err)
			return
		}
		dbUser := db.User{UserID: sessionUser.UserID}
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
