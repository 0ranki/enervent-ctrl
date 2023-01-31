package pingvinKL

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goburrow/modbus"
)

// single coil data
type pingvinCoil struct {
	Address     int    `json:"address"`
	Symbol      string `json:"symbol"`
	Value       bool   `json:"value"`
	Description string `json:"description"`
	Reserved    bool   `json:"reserved"`
}

func newCoil(address string, symbol string, description string) pingvinCoil {
	addr, err := strconv.Atoi(address)
	if err != nil {
		log.Fatal("newCoil: Atoi: ", err)
	}
	reserved := symbol == "-" && description == "-"
	coil := pingvinCoil{addr, symbol, false, description, reserved}
	return coil
}

// unit modbus data
type PingvinKL struct {
	Coils   []pingvinCoil
	buslock *sync.Mutex
}

// read a CSV file containing data for coils or registers
func readCsvLines(file string) [][]string {
	delim := ";"
	data := [][]string{}
	csv, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer csv.Close()
	scanner := bufio.NewScanner(csv)
	for scanner.Scan() {
		elements := strings.Split(scanner.Text(), delim)
		data = append(data, elements)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return data
}

// Configure the modbus parameters
func (p PingvinKL) getHandler() *modbus.RTUClientHandler {
	// TODO: read configuration from file, hardcoded for now
	handler := modbus.NewRTUClientHandler("/dev/ttyS0")
	handler.BaudRate = 19200
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 1500 * time.Millisecond
	return handler
}

func (p PingvinKL) Update() {
	handler := p.getHandler()
	p.buslock.Lock()
	err := handler.Connect()
	if err != nil {
		log.Fatal("Update: handler.Connect: ", err)
	}
	defer handler.Close()
	client := modbus.NewClient(handler)
	results, err := client.ReadCoils(0, uint16(len(p.Coils)))
	if err != nil {
		log.Fatal("Update: client.ReadCoils: ", err)
	}
	p.buslock.Unlock()
	// modbus.ReadCoils returns a byte array, with the first byte's bits representing coil values 0-7,
	// second byte coils 8-15 etc.
	// Within each byte, LSB represents the lowest n coil while MSB is the highest
	// e.g. reading the first 8 coils might return a byte array of length 1, with the following:
	// [4], which is 00000100, meaning all other coils are 0 except coil #2 (3rd coil)
	//
	k := 0                              // pingvinCoil index
	for i := 0; i < len(results); i++ { // loop through the byte array
		for j := 0; j < 8; j++ {
			// Here we loop through each bit in the byte, shifting right
			// and checking if the LSB after the shift is 1 with a bitwise AND
			// A coil value of 1 means on/true/yes, so == 1 returns the bool value
			// for each coil
			p.Coils[k].Value = (results[i] >> j & 0x1) == 1
			k++
		}
	}
}

func (p PingvinKL) ReadCoil(n uint16) []byte {
	handler := p.getHandler()
	p.buslock.Lock()
	err := handler.Connect()
	if err != nil {
		log.Fatal("ReadCoil: handler.Connect: ", err)
	}
	defer handler.Close()
	client := modbus.NewClient(handler)
	results, err := client.ReadCoils(n, 1)
	p.buslock.Unlock()
	if err != nil {
		log.Fatal("ReadCoil: client.ReadCoils: ", err)
	}
	p.Coils[n].Value = results[0] == 1
	return results
}

// create a PingvinKL struct, read coils and registers from CSVs
func New() PingvinKL {
	pingvin := PingvinKL{}
	pingvin.buslock = &sync.Mutex{}
	coilData := readCsvLines("coils.csv")
	for i := 0; i < len(coilData); i++ {
		pingvin.Coils = append(pingvin.Coils, newCoil(coilData[i][0], coilData[i][1], coilData[i][2]))
	}
	return pingvin
}
