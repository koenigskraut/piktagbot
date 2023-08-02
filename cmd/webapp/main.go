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
	WebAppRoot     = "/"
	WebAppHashPath = "/hash"
	WebAppWsPath   = "/ws"
)

//go:embed index.html
var htmlPage []byte
var indexTemplate = template.Must(template.New("main").Parse(string(htmlPage)))

func main() {
	certFile := os.Getenv("CERT_FILE")
	keyFile := os.Getenv("KEY_FILE")
	appPort := os.Getenv("APP_PORT")
	domain := os.Getenv("DOMAIN")
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
	http.HandleFunc(WebAppHashPath, handleVerification(os.Getenv("BOT_TOKEN")))
	http.HandleFunc(WebAppWsPath, handleWS)

	if err := http.ListenAndServeTLS(address, certFile, keyFile, nil); err != nil {
		log.Fatal(err)
	}
}
