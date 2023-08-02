package main

import (
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"log"
	"net/http"
	"time"
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
		for _, r := range []string{"a", "b", "c", "d"} {
			if err := wsutil.WriteServerText(conn, []byte(r)); err != nil {
				log.Println(3, err)
			}
			time.Sleep(time.Second)
		}
	}()
}
