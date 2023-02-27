package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/0ranki/enervent-ctrl/enervent-ctrl-go/pingvinKL"
)

// Remember to dereference the symbolic links under ./static/html
// prior to building the binary e.g. by using tar

//go:embed static/html/*
var static embed.FS

var (
	version = "0.0.5"
	pingvin pingvinKL.PingvinKL
	DEBUG   = false
)

func coils(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pathparams := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/coils/"), "/")
	if len(pathparams[0]) == 0 {
		json.NewEncoder(w).Encode(pingvin.Coils)
	} else if len(pathparams[0]) > 0 && r.Method == "GET" && len(pathparams) < 2 { // && r.Method == "POST"
		intaddr, err := strconv.Atoi(pathparams[0])
		if err != nil {
			log.Println("ERROR: Could not parse coil address", pathparams[0])
			log.Println(err)
			return
		}
		pingvin.ReadCoil(uint16(intaddr))
		json.NewEncoder(w).Encode(pingvin.Coils[intaddr])
	} else if len(pathparams[0]) > 0 && r.Method == "POST" && len(pathparams) == 2 {
		intaddr, err := strconv.Atoi(pathparams[0])
		if err != nil {
			log.Println("ERROR: Could not parse coil address", pathparams[0])
			log.Println(err)
			return
		}
		boolval, err := strconv.ParseBool(pathparams[1])
		if err != nil {
			log.Println("ERROR: Could not parse coil value", pathparams[1])
			log.Println(err)
			return
		}
		pingvin.WriteCoil(uint16(intaddr), boolval)
		json.NewEncoder(w).Encode(pingvin.Coils[intaddr])
	}
}

// func singlecoil(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")

// }

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
	pingvin = pingvinKL.New(DEBUG)
	pingvin.Update()
	go pingvin.Monitor(2)
	listen()
}
