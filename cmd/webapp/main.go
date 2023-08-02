package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
)

const (
	WebAppRoot = "/"
)

//go:embed index.html
var htmlPage []byte

func main() {
	certFile := os.Getenv("CERT_FILE")
	keyFile := os.Getenv("KEY_FILE")
	address := fmt.Sprintf(":%s", os.Getenv("APP_PORT"))
	http.HandleFunc(WebAppRoot, func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(htmlPage); err != nil {
			log.Println(err)
		}
	})
	if err := http.ListenAndServeTLS(address, certFile, keyFile, nil); err != nil {
		log.Fatal(err)
	}
}
