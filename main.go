package main

import (
	"crypto/sha256"
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/0ranki/enervent-ctrl/pingvin"
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
	version      = "0.1.3"
	device       pingvin.Pingvin
	config       Conf
	usernamehash [32]byte
	passwordhash [32]byte
)

type Conf struct {
	SerialAddress  string `yaml:"serial_address"`
	Port           int    `yaml:"port"`
	SslCertificate string `yaml:"ssl_certificate"`
	SslPrivatekey  string `yaml:"ssl_privatekey"`
	DisableAuth    bool   `yaml:"disable_auth"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	Interval       int    `yaml:"interval"`
	EnableMetrics  bool   `yaml:"enable_metrics"`
	LogFile        string `yaml:"log_file"`
	LogAccess      bool   `yaml:"log_access"`
	Debug          bool   `yaml:"debug"`
	ReadOnly       bool   `yaml:"read_only"`
}

// Start the HTTP server
func serve(cert, key *string) {
	log.Println("Starting service")
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
	http.HandleFunc("/coils/", authHandler(http.StripPrefix("/coils/", htmlroot)))
	http.HandleFunc("/registers/", authHandler(http.StripPrefix("/registers/", htmlroot)))
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
	device.Debug.Println("Certificate:\n", string(pub))
	device.Debug.Println("Key:\n", string(priv))
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
	yamldata, err := os.ReadFile(conffile)
	if err != nil {
		log.Println("Configuration file", conffile, "not found")
		log.Println("Generating", conffile, "with default values")
		initDefaultConfig(confpath)
		if yamldata, err = os.ReadFile(conffile); err != nil {
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
		SerialAddress:  "/dev/ttyS0",
		Port:           8888,
		SslCertificate: confpath + "/certificate.pem",
		SslPrivatekey:  confpath + "/privatekey.pem",
		DisableAuth:    false,
		Username:       "pingvin",
		Password:       "enervent",
		Interval:       4,
		EnableMetrics:  false,
		LogAccess:      false,
		LogFile:        "",
		Debug:          false,
		ReadOnly:       false,
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

// Read configuration. CLI flags take precedence over configuration file
func configure() {
	log.Println("Reading configuration")
	parseConfigFile()
	debugflag := flag.Bool("debug", config.Debug, "Enable debug logging")
	intervalflag := flag.Int("interval", config.Interval, "Set the interval of background updates")
	logaccflag := flag.Bool("httplog", config.LogAccess, "Enable HTTP access logging")
	generatecert := flag.Bool("regenerate-certs", false, "Generate a new SSL certificate. A new one is generated on startup as `~/.config/enervent-ctrl/server.crt` if it doesn't exist.")
	certflag := flag.String("cert", config.SslCertificate, "Path to SSL public key to use for HTTPS")
	keyflag := flag.String("key", config.SslPrivatekey, "Path to SSL private key to use for HTTPS")
	noauthflag := flag.Bool("disable-auth", config.DisableAuth, "Disable HTTP basic authentication")
	usernflag := flag.String("username", config.Username, "Username for HTTP Basic Authentication")
	passwflag := flag.String("password", config.Password, "Password for HTTP Basic Authentication")
	promflag := flag.Bool("enable-metrics", config.EnableMetrics, "Enable the built-in Prometheus exporter")
	logflag := flag.String("logfile", config.LogFile, "Path to log file. Default is empty string, log to stdout")
	serialflag := flag.String("serial", config.SerialAddress, "Path to serial console for RS-485 connection. Defaults to /dev/ttyS0")
	readOnly := flag.Bool("read-only", config.ReadOnly, "Read only mode, no writes to device are allowed")
	// TODO: log file flag
	flag.Parse()
	config.Debug = *debugflag
	config.Interval = *intervalflag
	config.LogAccess = *logaccflag
	config.SslCertificate = *certflag
	config.SslPrivatekey = *keyflag
	config.DisableAuth = *noauthflag
	config.Username = *usernflag
	config.Password = *passwflag
	config.EnableMetrics = *promflag
	config.LogFile = *logflag
	config.SerialAddress = *serialflag
	config.ReadOnly = *readOnly
	usernamehash = sha256.Sum256([]byte(config.Username))
	passwordhash = sha256.Sum256([]byte(config.Password))
	if len(config.LogFile) != 0 {
		logfile, err := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
		if err != nil {
			log.Fatal("Failed to open log file", config.LogFile)
		}
		log.SetOutput(logfile)
		log.Println("Opened logfile")
	}
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
		log.Println("Prometheus exporter enabled (/metrics)")
		prometheus.MustRegister(&device)
	}
}

func main() {
	log.Println("enervent-ctrl version", version)
	configure()
	device = *pingvin.New(config.SerialAddress, config.Debug)
	device.Update()
	go device.Monitor(config.Interval)
	serve(&config.SslCertificate, &config.SslPrivatekey)
	device.Quit()
}
