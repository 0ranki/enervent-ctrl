package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/0ranki/enervent-ctrl/enervent-ctrl-go/pingvinKL"
)

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
	static := http.FileServer(http.Dir("./static/html"))
	http.Handle("/", static)
	err := http.ListenAndServe(":8888", nil)
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
