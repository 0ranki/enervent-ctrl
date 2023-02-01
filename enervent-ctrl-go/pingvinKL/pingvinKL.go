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

// unit modbus data
type PingvinKL struct {
	Coils     []pingvinCoil
	Registers []pingvinRegister
	buslock   *sync.Mutex
}

// single register data
type pingvinRegister struct {
	Address     int    `json:"address"`
	Symbol      string `json:"symbol"`
	Value       int    `json:"value"`
	Signed      bool   `json:"signed"`
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

func newRegister(address string, symbol string, signed bool, description string) pingvinRegister {
	addr, err := strconv.Atoi(address)
	if err != nil {
		log.Fatal("newRegister: Atio: ")
	}
	reserved := symbol == "Reserved" && description == "Reserved"
	register := pingvinRegister{addr, symbol, 0, signed, description, reserved}
	return register
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

func (p PingvinKL) updateCoils() {
	handler := p.getHandler()
	p.buslock.Lock()
	err := handler.Connect()
	if err != nil {
		log.Fatal("updateCoils: handler.Connect: ", err)
	}
	defer handler.Close()
	client := modbus.NewClient(handler)
	results, err := client.ReadCoils(0, uint16(len(p.Coils)))
	if err != nil {
		log.Fatal("updateCoils: client.ReadCoils: ", err)
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

func (p PingvinKL) updateRegisters() {
	handler := p.getHandler()
	p.buslock.Lock()
	err := handler.Connect()
	if err != nil {
		log.Fatal("updateRegisters: handler.Connect: ", err)
	}
	defer handler.Close()
	client := modbus.NewClient(handler)
	regs := len(p.Registers)
	k := 0
	// modbus.ReadHoldingRegisters can read 125 regs at a time, so first we loop
	// until all the values are fethed, increasing the value of k for each register
	// When there are less than 125 registers to go, it's the last pass
	for k < regs {
		r := 125
		if regs-k < 125 {
			r = regs - k
		}
		results, err := client.ReadHoldingRegisters(uint16(k), uint16(r))
		if err != nil {
			log.Fatal("updateRegisters: client.ReadCoils: ", err)
		}
		// The values represent 16 bit integers, but modbus works with bytes
		// Each even byte of the returned []byte is the 8 MSBs of a new 16-bit
		// value, so for each even byte in the reponse slice we bitshift the byte
		// left by 8, then add the odd byte as is to the shifted 16-bit value
		msb := true
		value := 0
		for i := 0; i < len(results); i++ {
			if msb {
				value = int(results[i]) << 8
			} else {
				value += int(results[i])
				p.Registers[k].Value = value
				k++
			}
			msb = !msb
		}
	}
	p.buslock.Unlock()
}

func (p PingvinKL) Update() {
	p.updateCoils()
	p.updateRegisters()
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
	log.Println("Parsing coil data...")
	coilData := readCsvLines("coils.csv")
	for i := 0; i < len(coilData); i++ {
		pingvin.Coils = append(pingvin.Coils, newCoil(coilData[i][0], coilData[i][1], coilData[i][2]))
	}
	log.Println("Parsed", len(pingvin.Coils), "coils")
	log.Println("Parsing register data...")
	registerData := readCsvLines("registers.csv")
	for i := 0; i < len(registerData); i++ {
		signed := registerData[i][2] == "int16"
		pingvin.Registers = append(pingvin.Registers,
			newRegister(registerData[i][0], registerData[i][1], signed, registerData[i][6]))
	}
	log.Println("Parsed", len(pingvin.Registers), "registers")
	return pingvin
}
