package main

import (
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/0ranki/enervent-ctrl/enervent-ctrl-go/pingvinKL"
	"github.com/gorilla/handlers"
)

// Remember to dereference the symbolic links under ./static/html
// prior to building the binary e.g. by using tar

//go:embed static/html/*
var static embed.FS

var (
	version    = "0.0.11"
	pingvin    pingvinKL.PingvinKL
	DEBUG      *bool
	INTERVAL   *int
	ACCESS_LOG *bool
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
	} else if len(pathparams[0]) > 0 && r.Method == "POST" && len(pathparams) == 1 {
		intaddr, err := strconv.Atoi(pathparams[0])
		if err != nil {
			log.Println("ERROR: Could not parse coil address", pathparams[0])
			log.Println(err)
			return
		}
		pingvin.WriteCoil(uint16(intaddr), !pingvin.Coils[intaddr].Value)
		json.NewEncoder(w).Encode(pingvin.Coils[intaddr])
	}
}

func registers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pathparams := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/registers/"), "/")
	if len(pathparams[0]) == 0 {
		json.NewEncoder(w).Encode(pingvin.Registers)
	} else if len(pathparams[0]) > 0 && r.Method == "GET" && len(pathparams) < 2 { // && r.Method == "POST"
		intaddr, err := strconv.Atoi(pathparams[0])
		if err != nil {
			log.Println("ERROR: Could not parse register address", pathparams[0])
			log.Println(err)
			return
		}
		pingvin.ReadRegister(uint16(intaddr))
		json.NewEncoder(w).Encode(pingvin.Registers[intaddr])
	} else if len(pathparams[0]) > 0 && r.Method == "POST" && len(pathparams) == 2 {
		intaddr, err := strconv.Atoi(pathparams[0])
		if err != nil {
			log.Println("ERROR: Could not parse register address", pathparams[0])
			log.Println(err)
			return
		}
		intval, err := strconv.Atoi(pathparams[1])
		if err != nil {
			log.Println("ERROR: Could not parse register value", pathparams[1])
			log.Println(err)
			return
		}
		_, err = pingvin.WriteRegister(uint16(intaddr), uint16(intval))
		if err != nil {
			log.Println(err)
		}
		json.NewEncoder(w).Encode(pingvin.Registers[intaddr])
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
	if *ACCESS_LOG {
		handler := handlers.LoggingHandler(os.Stdout, http.DefaultServeMux)
		err = http.ListenAndServe(":8888", handler)
	} else {
		err = http.ListenAndServe(":8888", nil)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func configure() {
	log.Println("Reading configuration")
	DEBUG = flag.Bool("debug", false, "Enable debug logging")
	INTERVAL = flag.Int("interval", 4, "Set the interval of background updates")
	ACCESS_LOG = flag.Bool("httplog", false, "Enable HTTP access logging")
	flag.Parse()
	if *DEBUG {
		log.Println("Debug logging enabled")
	}
	if *ACCESS_LOG {
		log.Println("HTTP Access logging enabled")
	}
	log.Println("Update interval set to", *INTERVAL, "seconds")
}

func main() {
	log.Println("enervent-ctrl version", version)
	configure()
	pingvin = pingvinKL.New(*DEBUG)
	pingvin.Update()
	go pingvin.Monitor(*INTERVAL)
	listen()
}
