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
)

func coils(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pingvin.Coils)
}

func listen() {
	log.Println("Starting pingvinAPI...")
	http.HandleFunc("/api/v1/coils/", coils)
	log.Fatal(http.ListenAndServe(":8888", nil))
}

func main() {
	log.Println("enervent-ctrl version", version)
	pingvin = pingvinKL.New()
	pingvin.Update()
	listen()
}
