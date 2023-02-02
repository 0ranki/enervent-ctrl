package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"

	"github.com/0ranki/enervent-ctrl/enervent-ctrl-go/pingvinKL"
)

//go:embed static/html/*
var static embed.FS

var (
	version = "0.0.2"
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

func listen() {
	log.Println("Starting pingvinAPI...")
	http.HandleFunc("/api/v1/coils/", coils)
	http.HandleFunc("/api/v1/registers/", registers)
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
	listen()
}
