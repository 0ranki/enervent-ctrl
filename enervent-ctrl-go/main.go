package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"

	"github.com/0ranki/enervent-ctrl/enervent-ctrl-go/pingvinKL"
)

// Remember to dereference the symbolic links under ./static/html
// prior to building the binary e.g. by using tar

//go:embed static/html/*
var static embed.FS

var (
	version = "0.0.4"
	pingvin pingvinKL.PingvinKL
	DEBUG   = false
)

func coils(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pingvin.Coils)
}

func registers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if DEBUG {
		log.Println("Received request for /registers")
	}
	json.NewEncoder(w).Encode(pingvin.Registers)
	if DEBUG {
		log.Println("Handled request for /registers")
	}
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pingvin.Status)
}

func listen() {
	log.Println("Starting pingvinAPI...")
	http.HandleFunc("/api/v1/coils/", coils)
	http.HandleFunc("/api/v1/registers/", registers)
	http.HandleFunc("/api/v1/status", status)
	html, err := fs.Sub(static, "static/html")
	if err != nil {
		log.Fatal(err)
	}
	htmlroot := http.FileServer(http.FS(html))
	http.Handle("/", htmlroot)
	err = http.ListenAndServe(":8888", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Println("enervent-ctrl version", version)
	pingvin = pingvinKL.New()
	pingvin.Update()
	go pingvin.Monitor(15)
	listen()
}
