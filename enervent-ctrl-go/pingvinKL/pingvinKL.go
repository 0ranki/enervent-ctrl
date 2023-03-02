package pingvinKL

import (
	"bufio"
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
	debug      bool
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

type pingvinMeasurements struct {
	Roomtemp1       float32 `json:"room_temp1"`        // Room temperature at panel 1
	SupplyHeated    float32 `json:"supply_heated"`     // Temperature of supply air after heating
	SupplyHrc       float32 `json:"supply_hrc"`        // Temperature of supply air after heat recovery
	SupplyIntake    float32 `json:"supply_intake"`     // Temperature of outside air at device
	SupplyIntake24h float32 `json:"supply_intake_24h"` // 24h avg of outside air humidity
	SupplyHum       float32 `json:"supply_hum"`        // Supply air humidity
	Watertemp       float32 `json:"watertemp"`         // Heater element return water temperature
	ExtractIntake   float32 `json:"extract_intake"`    // Temperature of extract air
	ExtractHrc      float32 `json:"extract_hrc"`       // Temperature of extract air after heat recovery
	ExtractHum      float32 `json:"extract_hum"`       // Relative humidity of extract air
	ExtractHum48h   float32 `json:"extract_hum_48h"`   // 48h avg extract air humidity
}

type pingvinStatus struct {
	HeaterPct        int                 `json:"heater_pct"`         // After heater valve position
	HrcPct           int                 `json:"hrc_pct"`            // Heat recovery turn speed
	TempSetting      float32             `json:"temp_setting"`       // Requested room temperature
	FanPct           int                 `json:"fan_pct"`            // Circulation fan setting
	Measurements     pingvinMeasurements `json:"measurements"`       // Measurements
	HrcEffIn         int                 `json:"hrc_efficiency_in"`  // Calculated HRC efficiency, intake
	HrcEffEx         int                 `json:"hrc_efficiency_ex"`  // Calculated HRC efficiency, extract
	OpMode           string              `json:"op_mode"`            // Current operating mode, text representation
	DaysUntilService int                 `json:"days_until_service"` // Days until next filter service
	Uptime           string              `json:"uptime"`             // Unit uptime
	SystemTime       string              `json:"system_time"`        // Time and date in unit
}

var (
	// Mutually exclusive coils
	// Thanks to https://github.com/Jalle19/eda-modbus-bridge
	// 1 = Away mode
	// 2 = Away long mode
	// 3 = Overpressure
	// 6 = Max heating
	// 7 = Max cooling
	// 10 = Manual boost
	// 40 = Eco mode
	// Only one of these should be enabled at a time

	mutexcoils = []uint16{1, 2, 3, 6, 7, 10, 40}
)

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
	register := pingvinRegister{addr, symbol, 0, "0000000000000000", typ, description, reserved, multipl}
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
		results := []byte{}
		for retries := 0; retries < 5; retries++ {
			results, err = client.ReadHoldingRegisters(uint16(k), uint16(r))
			if len(results) > 0 {
				break
			} else if retries == 4 {
				log.Fatal("updateRegisters: client.ReadHoldingRegisters: ", err)
			} else if err != nil {
				log.Println("WARNING: updateRegisters: client.ReadHoldingRegisters: ", err)
			}
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
					// p.Registers[k].Bitfield = fmt.Sprintf("%16b", uvalue)
					p.Registers[k].Bitfield = ""
					for i := 16; i >= 0; i-- {
						x := 0
						if p.Registers[k].Value>>i&0x1 == 1 {
							x = 1
						}
						p.Registers[k].Bitfield = fmt.Sprintf("%s%s", p.Registers[k].Bitfield, strconv.Itoa(x))
					}
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

func (p *PingvinKL) WriteCoil(n uint16, val bool) bool {
	handler := p.getHandler()
	p.buslock.Lock()
	err := handler.Connect()
	if val {
		p.checkMutexCoils(n, handler)
	}
	if err != nil {
		log.Println("WARNING: WriteCoil: failed to connect handler")
		return false
	}
	defer handler.Close()
	var value uint16 = 0
	if val {
		value = 0xff00
	}
	client := modbus.NewClient(handler)
	results, err := client.WriteSingleCoil(n, value)
	p.buslock.Unlock()
	if err != nil {
		log.Println("ERROR: WriteCoil: ", err)
	}
	if (val && results[0] == 255) || (!val && results[0] == 0) {
		log.Println("WriteCoil: wrote coil", n, "to value", val)
	} else {
		log.Println("ERROR: WriteCoil: failed to write coil")
		return false

	}
	p.ReadCoil(n)
	return true
}

func (p *PingvinKL) WriteCoils(startaddr uint16, quantity uint16, vals []bool) error {
	handler := p.getHandler()
	p.buslock.Lock()
	err := handler.Connect()
	if err != nil {
		log.Println("WARNING: WriteCoils: failed to connect handler:", err)
		return err
	}
	defer handler.Close()
	p.updateCoils()
	coilslice := p.Coils[startaddr:(startaddr + quantity)]
	if len(coilslice) != len(vals) {
		return fmt.Errorf("ERROR: WriteCoils: vals ([]bool) is not the correct length")
	}
	// Convert slice of booleans to byte slice
	// representing individual bits
	// modbus.NewClient.WriteMultipleCoils wants the individual
	// bits in each byte "inverted", e.g. if you want to set 16 coils
	// with values 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, the
	// byte array needs to be [0x01,0x80] or [0b00000001, 0b10000000]
	bits := make([]byte, (len(coilslice)+7)/8)
	for i, coil := range coilslice {
		if coil.Value || vals[i] {
			// i/8 integer division, returns 0 for 0-7 etc.
			// i%8 loops through 0-7
			// If coil.Value or vals[i] is true, set i%8 + 1 least significant bit
			// to 1 in bits[i/8]
			// e.g. coil[19]:  (i/8 = 2, i%8 = 3)
			// -> bits[2] = (bits[2] | 0b00000001 << 3)
			// -> bits[2] = bits[2] | 0b00001000
			// -> 4th least sign. bit is set to 1
			bits[i/8] |= 0x01 << uint(i%8)
		}
		if !vals[i] {
			// bits contains the current values. If vals[i] is false,
			// the bit should be set to 0
			// ^(1 << 3) = ^0b00001000 = 0b11110111
			// 0b10101010 &| ^(1 << 3)
			//     0b10101010
			// AND 0b11110111
			// ->  0b10100010
			bits[i/8] &= ^(1 << uint(i%8))
		}
		if p.debug {
			log.Println("index:", i/8, "value:", bits[i/8], "shift:", i%8)
		}
	}
	log.Println(bits)
	client := modbus.NewClient(handler)
	results, err := client.WriteMultipleCoils(startaddr, quantity, bits)
	p.buslock.Unlock()
	if err != nil {
		log.Println("ERROR: WriteCoils: ", err)
		return err
	}
	log.Println(results)
	return nil
}

func (p *PingvinKL) checkMutexCoils(addr uint16, handler *modbus.RTUClientHandler) error {
	for _, mutexcoil := range mutexcoils {
		if mutexcoil == addr {
			for _, n := range mutexcoils {
				if p.Coils[n].Value {
					_, err := modbus.NewClient(handler).WriteSingleCoil(n, 0)
					if err != nil {
						log.Println("ERROR: checkMutexCoils:", err)
						return err
					}
				}
			}
			return nil
		}
	}
	return nil
}

func (p *PingvinKL) populateStatus() {
	hpct := p.Registers[49].Value / p.Registers[49].Multiplier
	if hpct > 100 {
		p.Status.HeaterPct = hpct - 100
		p.Status.HrcPct = 100
	} else {
		p.Status.HeaterPct = 0
		p.Status.HrcPct = hpct
	}
	p.Status.TempSetting = float32(p.Registers[135].Value) / float32(p.Registers[135].Multiplier)
	p.Status.FanPct = p.Registers[774].Value / p.Registers[774].Multiplier
	p.Status.Measurements.Roomtemp1 = float32(p.Registers[1].Value) / float32(p.Registers[1].Multiplier)
	p.Status.Measurements.SupplyHeated = float32(p.Registers[8].Value) / float32(p.Registers[8].Multiplier)
	p.Status.Measurements.SupplyHrc = float32(p.Registers[7].Value) / float32(p.Registers[7].Multiplier)
	p.Status.Measurements.SupplyIntake = float32(p.Registers[6].Value) / float32(p.Registers[6].Multiplier)
	p.Status.Measurements.SupplyIntake24h = float32(p.Registers[134].Value) / float32(p.Registers[134].Multiplier)
	p.Status.Measurements.SupplyHum = float32(p.Registers[36].Value) / float32(p.Registers[46].Multiplier)
	p.Status.Measurements.Watertemp = float32(p.Registers[12].Value) / float32(p.Registers[12].Multiplier)
	p.Status.Measurements.ExtractIntake = float32(p.Registers[10].Value) / float32(p.Registers[10].Multiplier)
	p.Status.Measurements.ExtractHrc = float32(p.Registers[9].Value) / float32(p.Registers[9].Multiplier)
	p.Status.Measurements.ExtractHum = float32(p.Registers[28].Value) / float32(p.Registers[28].Multiplier)
	p.Status.Measurements.ExtractHum48h = float32(p.Registers[50].Value) / float32(p.Registers[50].Multiplier)
	p.Status.HrcEffIn = p.Registers[29].Value / p.Registers[29].Multiplier
	p.Status.HrcEffEx = p.Registers[30].Value / p.Registers[30].Multiplier
	p.Status.OpMode = parseStatus(p.Registers[44].Value)
	// TODO: Alarms, n of alarms
	p.Status.DaysUntilService = p.Registers[538].Value / p.Registers[538].Multiplier
	// TODO: Uptime & date in separate functions
}

func parseStatus(value int) string {
	val := int16(value)
	pingvinStatuses := []string{
		"Max cooling",
		"Max heating",
		"Stopped by alarm",
		"Stopped by user",
		"Away",
		"reserved",
		"Adaptive",
		"CO2 boost",
		"RH boost",
		"Manual boost",
		"Overpressure",
		"Cooker hood mode",
		"Central vac mode",
		"Electric heater cooloff",
		"Summer night cooling",
		"HRC defrost",
	}
	for i := 0; i < 15; i++ {
		if val>>i&0x1 == 1 {
			return pingvinStatuses[i]
		}
	}
	return "Normal"

}

func (p *PingvinKL) Monitor(interval int) {
	for {
		time.Sleep(time.Duration(interval) * time.Second)
		if p.debug {
			log.Println("DEBUG: Updating values")
		}
		p.Update()
		if p.debug {
			log.Println("DEBUG: coils:", p.Coils)
			log.Println("DEBUG: registers:", p.Registers)
		}
	}
}

// create a PingvinKL struct, read coils and registers from CSVs
func New(debug bool) PingvinKL {
	pingvin := PingvinKL{}
	pingvin.debug = debug
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
