package pingvinKL

import (
	"bufio"
	"encoding/json"
	"fmt"
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
	Coils      []pingvinCoil
	Registers  []pingvinRegister
	Status     pingvinStatus
	buslock    *sync.Mutex
	statuslock *sync.Mutex
}

// single register data
type pingvinRegister struct {
	Address     int    `json:"address"`
	Symbol      string `json:"symbol"`
	Value       int    `json:"value"`
	Bitfield    string `json:"bitfield"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Reserved    bool   `json:"reserved"`
	Multiplier  int    `json:"multiplier"`
}

type pingvinVentInfo struct {
	SupplyHeated    int `json:"supply_heated"`
	SupplyHrc       int `json:"supply_hrc"`
	SupplyIntake    int `json:"supply_intake"`
	SupplyIntake24h int `json:"supply_intake_24h"`
	SupplyHum       int `json:"supply_hum"`
	ExtractIntake   int `json:"extract_intake"`
	ExtractHrc      int `json:"extract_hrc"`
	ExtractHum      int `json:"extract_hum"`
	ExtractHum48h   int `json:"extract_hum_48h"`
}

type pingvinStatus struct {
	HeaterPct        int             `json:"heater_pct"`
	HrcPct           int             `json:"hrc_pct"`
	TempSetting      int             `json:"temp_setting"`
	FanPct           int             `json:"fan_pct"`
	VentInfo         pingvinVentInfo `json:"vent_info"`
	HrcEffIn         int             `json:"hrc_efficiency_in"`
	HrcEffEx         int             `json:"hrc_efficiency_ex"`
	OpMode           string          `json:"op_mode"`
	DaysUntilService int             `json:"days_until_service"`
	Uptime           string          `json:"uptime"`
	SystemTime       string          `json:"system_time"`
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

func newRegister(address, symbol, typ, multiplier, description string) pingvinRegister {
	addr, err := strconv.Atoi(address)
	if err != nil {
		log.Fatal("newRegister: Atoi(address): ", err)
	}
	multipl := 1
	if len(multiplier) > 0 {
		multipl, err = strconv.Atoi(multiplier)
		if err != nil {
			log.Fatal("newRegister: Atoi(multiplier): ", err)
		}
	}
	reserved := symbol == "Reserved" && description == "Reserved"
	register := pingvinRegister{addr, symbol, 0, "00000000", typ, description, reserved, multipl}
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

func (p *PingvinKL) updateCoils() {
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

func (p *PingvinKL) updateRegisters() {
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
		value := int16(0)
		uvalue := uint16(0)
		for i := 0; i < len(results); i++ {
			if msb {
				value = int16(results[i]) << 8
				uvalue = uint16(results[i]) << 8
			} else {
				value += int16(results[i])
				uvalue += uint16(results[i])
				if p.Registers[k].Type == "int16" {
					p.Registers[k].Value = int(value)
				}
				if p.Registers[k].Type == "uint16" || p.Registers[k].Type == "enumeration" {
					p.Registers[k].Value = int(uvalue)
				}
				if p.Registers[k].Type == "bitfield" {
					p.Registers[k].Value = int(value)
					p.Registers[k].Bitfield = fmt.Sprintf("%08b", uvalue)
				}
				k++
			}
			msb = !msb
		}
	}
	p.buslock.Unlock()
}

func (p *PingvinKL) Update() {
	p.updateCoils()
	p.updateRegisters()
	p.populateStatus()
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

func (p *PingvinKL) populateStatus() {
	hpct := p.Registers[49].Value / p.Registers[49].Multiplier
	log.Println(hpct)
	if hpct > 100 {
		p.Status.HeaterPct = hpct - 100
		p.Status.HrcPct = 100
	} else {
		p.Status.HeaterPct = 0
		p.Status.HrcPct = hpct
	}
	p.Status.TempSetting = p.Registers[135].Value / p.Registers[135].Multiplier
	p.Status.FanPct = p.Registers[774].Value / p.Registers[774].Multiplier
	p.Status.VentInfo.SupplyHeated = p.Registers[8].Value / p.Registers[8].Multiplier
	p.Status.VentInfo.SupplyHrc = p.Registers[7].Value / p.Registers[7].Multiplier
	p.Status.VentInfo.SupplyIntake = p.Registers[6].Value / p.Registers[6].Multiplier
	p.Status.VentInfo.SupplyIntake24h = p.Registers[134].Value / p.Registers[134].Multiplier
	p.Status.VentInfo.SupplyHum = p.Registers[36].Value / p.Registers[46].Multiplier
	p.Status.VentInfo.ExtractIntake = p.Registers[10].Value / p.Registers[10].Multiplier
	p.Status.VentInfo.ExtractHrc = p.Registers[9].Value / p.Registers[9].Multiplier
	p.Status.VentInfo.ExtractHum = p.Registers[28].Value / p.Registers[28].Multiplier
	p.Status.VentInfo.ExtractHum48h = p.Registers[50].Value / p.Registers[50].Multiplier
	p.Status.HrcEffIn = p.Registers[29].Value / p.Registers[29].Multiplier
	p.Status.HrcEffEx = p.Registers[30].Value / p.Registers[30].Multiplier
	// TODO: Operating mode in separate function
	// TODO: Alarms, n of alarms
	p.Status.DaysUntilService = p.Registers[538].Value / p.Registers[538].Multiplier
	// TODO: Uptime & date in separate functions
	json.NewEncoder(log.Writer()).Encode(p.Status)
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
		pingvin.Registers = append(pingvin.Registers,
			newRegister(registerData[i][0], registerData[i][1], registerData[i][2], registerData[i][3], registerData[i][6]))
	}
	log.Println("Parsed", len(pingvin.Registers), "registers")
	return pingvin
}
