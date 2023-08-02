package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

const (
	WebAppRoot         = "/"
	WebAppHashPath     = "/hash"
	WebAppWsPath       = "/ws"
	WebAppStickersPath = "/stickers/"
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

func main() {
	address := fmt.Sprintf(":%s", appPort)
	wsPath := fmt.Sprintf("wss://%s:%s%s", domain, appPort, WebAppWsPath)

	http.HandleFunc(WebAppRoot, func(w http.ResponseWriter, r *http.Request) {
		err := indexTemplate.Execute(w, struct {
			WSPath string
		}{wsPath})
		if err != nil {
			log.Println(err)
		}
	})
	http.HandleFunc(WebAppHashPath, handleVerification(botToken))
	http.HandleFunc(WebAppWsPath, handleWS)
	http.Handle(WebAppStickersPath, http.StripPrefix(WebAppStickersPath, http.FileServer(http.Dir(stickerPath))))

	if err := http.ListenAndServeTLS(address, certFile, keyFile, nil); err != nil {
		log.Fatal(err)
	}
}
