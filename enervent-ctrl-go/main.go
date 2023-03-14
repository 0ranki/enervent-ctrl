package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/0ranki/enervent-ctrl/enervent-ctrl-go/pingvinKL"
	"github.com/0ranki/https-go"
	"github.com/gorilla/handlers"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/yaml.v3"
)

// Remember to dereference the symbolic links under ./static/html
// prior to building the binary e.g. by using tar

//go:embed static/html/*
var static embed.FS

var (
	version      = "0.0.23"
	pingvin      pingvinKL.PingvinKL
	config       Conf
	usernamehash [32]byte
	passwordhash [32]byte
)

type Conf struct {
	Port           int    `yaml:"port"`
	SslCertificate string `yaml:"ssl_certificate"`
	SslPrivatekey  string `yaml:"ssl_privatekey"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	Interval       int    `yaml:"interval"`
	EnableMetrics  bool   `yaml:"enable_metrics"`
	LogAccess      bool   `yaml:"log_access"`
	Debug          bool   `yaml:"debug"`
}

// HTTP Basic Authentication middleware for http.HandlerFunc
func authHandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	// Based on https://www.alexedwards.net/blog/basic-authentication-in-go
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if ok {
			userHash := sha256.Sum256([]byte(user))
			passHash := sha256.Sum256([]byte(pass))
			usernameMatch := (subtle.ConstantTimeCompare(userHash[:], usernamehash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passHash[:], passwordhash[:]) == 1)
			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}
		if len(user) == 0 {
			user = "-"
		}
		log.Println("Authentication failed: IP:", r.RemoteAddr, "URI:", r.RequestURI, "username:", user)
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

// HTTP Basic Authentication middleware for http.Handler
func authHandler(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if ok {
			userHash := sha256.Sum256([]byte(user))
			passHash := sha256.Sum256([]byte(pass))
			usernameMatch := (subtle.ConstantTimeCompare(userHash[:], usernamehash[:]) == 1)
			passwordMatch := (subtle.ConstantTimeCompare(passHash[:], passwordhash[:]) == 1)
			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}
		if len(user) == 0 {
			user = "-"
		}
		log.Println("Authentication failed: IP:", r.RemoteAddr, "URI:", r.RequestURI, "username:", user)
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

// \/api/v1/coils endpoint
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

// \/api/v1/registers endpoint
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

// \/status endpoint
func status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pingvin.Status)
}

// \/api/v1/temperature endpoint
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

// Start the HTTP server
func serve(cert, key *string) {
	log.Println("Starting pingvinAPI...")
	http.HandleFunc("/api/v1/coils/", authHandlerFunc(coils))
	http.HandleFunc("/api/v1/status", authHandlerFunc(status))
	http.HandleFunc("/api/v1/registers/", authHandlerFunc(registers))
	http.HandleFunc("/api/v1/temperature/", authHandlerFunc(temperature))
	if config.EnableMetrics {
		http.Handle("/metrics", promhttp.Handler())
	}
	html, err := fs.Sub(static, "static/html")
	if err != nil {
		log.Fatal(err)
	}
	htmlroot := http.FileServer(http.FS(html))
	http.HandleFunc("/", authHandler(htmlroot))
	logdst, err := os.OpenFile(os.DevNull, os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Fatal(err)
	}
	if config.LogAccess {
		logdst = os.Stdout
	}
	handler := handlers.LoggingHandler(logdst, http.DefaultServeMux)
	err = http.ListenAndServeTLS(":8888", *cert, *key, handler)
	if err != nil {
		log.Fatal(err)
	}
}

// Generate self-signed SSL keypair
func generateCertificate(cert, key string) {
	opts := https.GenerateOptions{Host: "enervent-ctrl.local", RSABits: 4096, ValidFor: 10 * 365 * 24 * time.Hour}
	log.Println("Generating new self-signed SSL keypair")
	log.Println("This may take a while...")
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

// Read & parse the configuration file
func parseConfigFile() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Could not determine user home directory")
	}
	confpath := homedir + "/.config/enervent-ctrl"
	if _, err := os.Stat(confpath); err != nil {
		log.Println("Generating configuration directory", confpath)
		if err := os.MkdirAll(confpath, 0700); err != nil {
			log.Fatal("Failed to generate configuration directory:", err)
		}
	}
	conffile := confpath + "/configuration.yaml"
	yamldata, err := ioutil.ReadFile(conffile)
	if err != nil {
		log.Println("Configuration file", conffile, "not found")
		log.Println("Generating", conffile, "with default values")
		initDefaultConfig(confpath)
		if yamldata, err = ioutil.ReadFile(conffile); err != nil {
			log.Fatal("Error parsing configuration:", err)
		}
	}
	err = yaml.Unmarshal(yamldata, &config)
	if err != nil {
		log.Fatal("Failed to parse YAML:", err)
	}
}

// Write the default configuration to $HOME/.config/enervent-ctrl/configuration.yaml
func initDefaultConfig(confpath string) {
	config = Conf{
		8888,
		confpath + "/certificate.pem",
		confpath + "/privatekey.pem",
		"pingvin",
		"enervent",
		4,
		false,
		false,
		false,
	}
	conffile := confpath + "/configuration.yaml"
	confbytes, err := yaml.Marshal(&config)
	if err != nil {
		log.Println("Error writing default configuration:", err)
	}
	if err := os.WriteFile(conffile, confbytes, 0600); err != nil {
		log.Fatal("Failed to write default configuration:", err)
	}
}

// Read configuration. CLI flags take presedence over configuration file
func configure() {
	log.Println("Reading configuration")
	parseConfigFile()
	debugflag := flag.Bool("debug", config.Debug, "Enable debug logging")
	intervalflag := flag.Int("interval", config.Interval, "Set the interval of background updates")
	logaccflag := flag.Bool("httplog", config.LogAccess, "Enable HTTP access logging")
	generatecert := flag.Bool("regenerate-certs", false, "Generate a new SSL certificate. A new one is generated on startup as `~/.config/enervent-ctrl/server.crt` if it doesn't exist.")
	certflag := flag.String("cert", config.SslCertificate, "Path to SSL public key to use for HTTPS")
	keyflag := flag.String("key", config.SslPrivatekey, "Path to SSL private key to use for HTTPS")
	usernflag := flag.String("username", config.Username, "Username for HTTP Basic Authentication")
	passwflag := flag.String("password", config.Password, "Password for HTTP Basic Authentication")
	promflag := flag.Bool("enable-metrics", config.EnableMetrics, "Enable the built-in Prometheus exporter")
	// TODO: log file flag
	flag.Parse()
	config.Debug = *debugflag
	config.Interval = *intervalflag
	config.LogAccess = *logaccflag
	config.SslCertificate = *certflag
	config.SslPrivatekey = *keyflag
	config.Username = *usernflag
	config.Password = *passwflag
	config.EnableMetrics = *promflag
	usernamehash = sha256.Sum256([]byte(config.Username))
	passwordhash = sha256.Sum256([]byte(config.Password))
	// Check that certificate file exists, generate if needed
	if _, err := os.Stat(config.SslCertificate); err != nil || *generatecert {
		generateCertificate(config.SslCertificate, config.SslPrivatekey)
	}
	// Enable debug if configured
	if config.Debug {
		log.Println("Debug logging enabled")
	}
	// Enable HTTP access logging if configured
	if config.LogAccess {
		log.Println("HTTP Access logging enabled")
	}
	log.Println("Update interval set to", config.Interval, "seconds")
	if config.EnableMetrics {
		prometheus.MustRegister(&pingvin)
	}
}

func main() {
	log.Println("enervent-ctrl version", version)
	configure()
	pingvin = pingvinKL.New(config.Debug)
	pingvin.Update()
	go pingvin.Monitor(config.Interval)
	serve(&config.SslCertificate, &config.SslPrivatekey)
	pingvin.Quit()
}
