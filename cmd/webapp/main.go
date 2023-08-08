package main

import (
	"context"
	_ "embed"
	"fmt"
	db "github.com/koenigskraut/piktagbot/database"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	WebAppRoot         = "/"
	WebAppHashPath     = "/hash"
	WebAppWsPath       = "/ws"
	WebAppStickersPath = "/stickers/"
	WebAppOrderUpdate  = "/updateOrder"
)

//go:embed index.html
var htmlPage []byte
var indexTemplate = template.Must(template.New("main").Parse(string(htmlPage)))

var (
	certFile    = os.Getenv("CERT_FILE")
	keyFile     = os.Getenv("KEY_FILE")
	appPort     = os.Getenv("APP_PORT")
	domain      = os.Getenv("DOMAIN")
	stickerPath = os.Getenv("STICKER_PATH")
	botToken    = os.Getenv("BOT_TOKEN")
)

func init() {
	downloaded.init(stickerPath)
}

func main() {
	address := fmt.Sprintf(":%s", appPort)
	wsPath := fmt.Sprintf("wss://%s:%s%s", domain, appPort, WebAppWsPath)

	db.InitializeDB()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go initializeTelegram(ctx)

	http.HandleFunc(WebAppRoot, func(w http.ResponseWriter, r *http.Request) {
		err := indexTemplate.Execute(w, struct {
			WSPath string
		}{wsPath})
		if err != nil {
			log.Println(err)
		}
	})
	http.HandleFunc(WebAppHashPath, handleVerification)
	http.HandleFunc(WebAppWsPath, handleWS)
	//http.Handle(WebAppStickersPath, http.StripPrefix(WebAppStickersPath, http.FileServer(http.Dir(stickerPath))))
	http.HandleFunc(WebAppOrderUpdate, handleOrderUpdate)

	go func() {
		err := http.ListenAndServeTLS(address, certFile, keyFile, nil)
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}()
	<-ctx.Done()

	// hacky way to wait until telegram in goroutine reports closing
	time.Sleep(100 * time.Millisecond)
}
