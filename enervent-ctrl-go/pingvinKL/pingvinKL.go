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

type pingvinVentInfo struct {
	Roomtemp1       int `json:"room_temp1"`        // Room temperature at panel 1
	SupplyHeated    int `json:"supply_heated"`     // Temperature of supply air after heating
	SupplyHrc       int `json:"supply_hrc"`        // Temperature of supply air after heat recovery
	SupplyIntake    int `json:"supply_intake"`     // Temperature of outside air at device
	SupplyIntake24h int `json:"supply_intake_24h"` // 24h avg of outside air humidity
	SupplyHum       int `json:"supply_hum"`        // Supply air humidity
	ExtractIntake   int `json:"extract_intake"`    // Temperature of extract air
	ExtractHrc      int `json:"extract_hrc"`       // Temperature of extract air after heat recovery
	ExtractHum      int `json:"extract_hum"`       // Relative humidity of extract air
	ExtractHum48h   int `json:"extract_hum_48h"`   // 48h avg extract air humidity
}

type pingvinStatus struct {
	HeaterPct        int             `json:"heater_pct"`         // After heater valve position
	HrcPct           int             `json:"hrc_pct"`            // Heat recovery turn speed
	TempSetting      int             `json:"temp_setting"`       // Requested room temperature
	FanPct           int             `json:"fan_pct"`            // Circulation fan setting
	VentInfo         pingvinVentInfo `json:"vent_info"`          // Measurements
	HrcEffIn         int             `json:"hrc_efficiency_in"`  // Calculated HRC efficiency, intake
	HrcEffEx         int             `json:"hrc_efficiency_ex"`  // Calculated HRC efficiency, extract
	OpMode           string          `json:"op_mode"`            // Current operating mode, text representation
	DaysUntilService int             `json:"days_until_service"` // Days until next filter service
	Uptime           string          `json:"uptime"`             // Unit uptime
	SystemTime       string          `json:"system_time"`        // Time and date in unit
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

func (p *PingvinKL) populateStatus() {
	hpct := p.Registers[49].Value / p.Registers[49].Multiplier
	if hpct > 100 {
		p.Status.HeaterPct = hpct - 100
		p.Status.HrcPct = 100
	} else {
		p.Status.HeaterPct = 0
		p.Status.HrcPct = hpct
	}
	p.Status.TempSetting = p.Registers[135].Value / p.Registers[135].Multiplier
	p.Status.FanPct = p.Registers[774].Value / p.Registers[774].Multiplier
	p.Status.VentInfo.Roomtemp1 = p.Registers[1].Value / p.Registers[1].Multiplier
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
	p.Status.OpMode = parseStatus(p.Registers[44].Value)
	// TODO: Alarms, n of alarms
	p.Status.DaysUntilService = p.Registers[538].Value / p.Registers[538].Multiplier
	// TODO: Uptime & date in separate functions
}

func parseStatus(value int) string {
	val := int16(value)
	pingvinStatuses := []string{
		"Normal",
		"Max heating",
		"Max cooling",
		"Stopped by alarm",
		"Stopped by user",
		"Away",
		"reserved",
		"Temperature boost",
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
	for i := 1; i <= 16; i++ {
		if val>>i&0x1 == 1 {
			return pingvinStatuses[i]
		}
	}
	return "Normal"

}

func (p *PingvinKL) Monitor(interval int) {
	for {
		time.Sleep(time.Duration(interval) * time.Second)
		p.Update()
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
