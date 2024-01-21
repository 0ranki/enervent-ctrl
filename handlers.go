package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// HTTP Basic Authentication middleware for http.HandlerFunc
// This is used for the API
func authHandlerFunc(next http.HandlerFunc) http.HandlerFunc {
	// Based on https://www.alexedwards.net/blog/basic-authentication-in-go
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.DisableAuth {
			next.ServeHTTP(w, r)
			return
		}
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
// Used for the HTML monitor views
func authHandler(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.DisableAuth {
			next.ServeHTTP(w, r)
			return
		}
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

// /api/v1/coils endpoint
func coils(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pathparams := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/coils/"), "/")
	if len(pathparams[0]) == 0 {
		_ = json.NewEncoder(w).Encode(device.Coils)
	} else if len(pathparams[0]) > 0 && r.Method == "GET" && len(pathparams) < 2 { // && r.Method == "POST"
		intaddr, err := strconv.Atoi(pathparams[0])
		if err != nil {
			log.Println("ERROR: Could not parse coil address", pathparams[0])
			log.Println(err)
			return
		}
		err = device.ReadCoil(uint16(intaddr))
		if err != nil {
			log.Println("ERROR ReadCoil: client.ReadCoils: ", err)
		}
		_ = json.NewEncoder(w).Encode(device.Coils[intaddr])
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
		if config.ReadOnly {
			log.Println("WARNING: Read only mode, refusing to write to device")
		} else {
			device.WriteCoil(uint16(intaddr), boolval)
		}
		_ = json.NewEncoder(w).Encode(device.Coils[intaddr])
	} else if len(pathparams[0]) > 0 && r.Method == "POST" && len(pathparams) == 1 {
		intaddr, err := strconv.Atoi(pathparams[0])
		if err != nil {
			log.Println("ERROR: Could not parse coil address", pathparams[0])
			log.Println(err)
			return
		}
		if config.ReadOnly {
			log.Println("WARNING: Read only mode, refusing to write to device")
		} else {
			device.WriteCoil(uint16(intaddr), !device.Coils[intaddr].Value)
		}
		_ = json.NewEncoder(w).Encode(device.Coils[intaddr])
	}
}

// /api/v1/registers endpoint
func registers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pathparams := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/registers/"), "/")
	if len(pathparams[0]) == 0 {
		_ = json.NewEncoder(w).Encode(device.Registers)
	} else if len(pathparams[0]) > 0 && r.Method == "GET" && len(pathparams) < 2 { // && r.Method == "POST"
		intaddr, err := strconv.Atoi(pathparams[0])
		if err != nil {
			log.Println("ERROR: Could not parse register address", pathparams[0])
			log.Println(err)
			return
		}
		_, err = device.ReadRegister(uint16(intaddr))
		if err != nil {
			log.Println("ERROR: ReadRegister:", err)
		}
		_ = json.NewEncoder(w).Encode(device.Registers[intaddr])
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
		if config.ReadOnly {
			log.Println("WARNING: Read only mode, refusing to write to device")
		} else {
			_, err = device.WriteRegister(uint16(intaddr), uint16(intval))
			if err != nil {
				log.Println(err)
			}
		}
		_ = json.NewEncoder(w).Encode(device.Registers[intaddr])
	}
}

// /status endpoint
func status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(device.Status)
}

// /api/v1/temperature endpoint
func temperature(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pathparams := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/v1/temperature/"), "/")
	if len(pathparams[0]) > 0 && r.Method == "POST" && len(pathparams) == 1 {
		err := device.Temperature(pathparams[0])
		if err != nil {
			log.Println("ERROR: ", err)
		}
		_ = json.NewEncoder(w).Encode(device.Registers[135])
	} else {
		return
	}
}
