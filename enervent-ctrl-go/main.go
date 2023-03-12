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
	"time"

	"github.com/0ranki/enervent-ctrl/enervent-ctrl-go/pingvinKL"
	"github.com/gorilla/handlers"
	"github.com/rocketlaunchr/https-go"
)

// Remember to dereference the symbolic links under ./static/html
// prior to building the binary e.g. by using tar

//go:embed static/html/*
var static embed.FS

var (
	version    = "0.0.18"
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

func temperature(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pathparams := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/temperature/"), "/")
	if len(pathparams[0]) > 0 && r.Method == "POST" && len(pathparams) == 1 {
		pingvin.Temperature(pathparams[0])
		json.NewEncoder(w).Encode(pingvin.Registers[135])
	} else {
		return
	}
}

func listen() {
	log.Println("Starting pingvinAPI...")
	http.HandleFunc("/api/v1/coils/", coils)
	http.HandleFunc("/api/v1/registers/", registers)
	http.HandleFunc("/api/v1/status", status)
	http.HandleFunc("/api/v1/temperature/", temperature)
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

func generateCertificate(certpath, cert, key string) {
	if _, err := os.Stat(certpath); err != nil {
		log.Println("Generating configuration directory", certpath)
		if err := os.MkdirAll(certpath, 0750); err != nil {
			log.Fatal("Failed to generate configuration directory:", err)
		}
	}
	opts := https.GenerateOptions{Host: "enervent-ctrl.local", RSABits: 4096, ValidFor: 10 * 365 * 24 * time.Hour}
	log.Println("Generating new self-signed SSL keypair to ", certpath)
	pub, priv, err := https.GenerateKeys(opts)
	if err != nil {
		log.Fatal("Error generating SSL certificate: ", err)
	}
	pingvin.Debug.Println("Certificate:\n", string(pub))
	pingvin.Debug.Println("Key:\n", string(priv))
	if err := os.WriteFile(key, priv, 0600); err != nil {
		log.Fatal("Error writing private key ", key, ": ", err)
	}
	log.Println("Wrote new SSL private key ", cert)
	if err := os.WriteFile(cert, pub, 0644); err != nil {
		log.Fatal("Error writing certificate ", cert, ": ", err)
	}
	log.Println("Wrote new SSL public key ", cert)
}

func configure() {
	log.Println("Reading configuration")
	DEBUG = flag.Bool("debug", false, "Enable debug logging")
	INTERVAL = flag.Int("interval", 4, "Set the interval of background updates")
	ACCESS_LOG = flag.Bool("httplog", false, "Enable HTTP access logging")
	generatecert := flag.Bool("regenerate-certs", false, "Generate a new SSL certificate. A new one is generated on startup as `~/.config/enervent-ctrl/server.crt` if it doesn't exist.")
	// TODO: flag for cerificate path
	// TODO: log file flag
	flag.Parse()
	// Get the user home directory path
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Could not determine user home directory")
	}
	certpath := homedir + "/.config/enervent-ctrl/"
	cert := certpath + "certificate.pem"
	key := certpath + "privatekey.pem"
	// Check that certificate file exists
	if _, err = os.Stat(cert); err != nil || *generatecert {
		generateCertificate(certpath, cert, key)
	}
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
	pingvin.Quit()
}
