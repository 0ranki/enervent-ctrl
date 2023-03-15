package pingvin

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
	"github.com/prometheus/client_golang/prometheus"
)

// single coil data
type pingvinCoil struct {
	Address     int              `json:"address"`
	Symbol      string           `json:"symbol"`
	Value       bool             `json:"value"`
	Description string           `json:"description"`
	Reserved    bool             `json:"reserved"`
	PromDesc    *prometheus.Desc `json:"-"`
}

// unit modbus data
type PingvinKL struct {
	Coils        []pingvinCoil
	Registers    []pingvinRegister
	Status       pingvinStatus
	buslock      *sync.Mutex
	statuslock   *sync.Mutex
	handler      *modbus.RTUClientHandler
	modbusclient modbus.Client
	Debug        PingvinLogger
}

// single register data
type pingvinRegister struct {
	Address     int              `json:"address"`
	Symbol      string           `json:"symbol"`
	Value       int              `json:"value"`
	Bitfield    string           `json:"bitfield"`
	Type        string           `json:"type"`
	Description string           `json:"description"`
	Reserved    bool             `json:"reserved"`
	Multiplier  int              `json:"multiplier"`
	PromDesc    *prometheus.Desc `json:"-"`
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
	HeaterPct    int                 `json:"heater_pct"`        // After heater valve position
	HrcPct       int                 `json:"hrc_pct"`           // Heat recovery turn speed
	TempSetting  float32             `json:"temp_setting"`      // Requested room temperature
	FanPct       int                 `json:"fan_pct"`           // Circulation fan setting
	Measurements pingvinMeasurements `json:"measurements"`      // Measurements
	HrcEffIn     int                 `json:"hrc_efficiency_in"` // Calculated HRC efficiency, intake
	HrcEffEx     int                 `json:"hrc_efficiency_ex"` // Calculated HRC efficiency, extract
	OpMode       string              `json:"op_mode"`           // Current operating mode, text representation
	Uptime       string              `json:"uptime"`            // Unit uptime
	SystemTime   string              `json:"system_time"`       // Time and date in unit
	Coils        []pingvinCoil       `json:"coils"`
}

type PingvinLogger struct {
	dbg bool
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

func (logger *PingvinLogger) Println(msg ...any) {
	if logger.dbg {
		log.Println(msg...)
	}
}

func newCoil(address string, symbol string, description string) pingvinCoil {
	addr, err := strconv.Atoi(address)
	if err != nil {
		log.Fatal("newCoil: Atoi: ", err)
	}
	reserved := symbol == "-" && description == "-"
	if !reserved {
		promdesc := strings.ToLower(symbol)
		zpadaddr := fmt.Sprintf("%02d", addr)
		promdesc = strings.Replace(promdesc, "_", "_"+zpadaddr+"_", 1)
		return pingvinCoil{addr, symbol, false, description, reserved,
			prometheus.NewDesc(
				prometheus.BuildFQName("", "pingvin", promdesc),
				description,
				nil,
				nil,
			),
		}
	}
	return pingvinCoil{addr, symbol, false, description, reserved, nil}
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

	if !reserved {
		promdesc := strings.ToLower(symbol)
		zpadaddr := fmt.Sprintf("%03d", addr)
		promdesc = strings.Replace(promdesc, "_", "_"+zpadaddr+"_", 1)
		return pingvinRegister{
			addr,
			symbol,
			0,
			"0000000000000000",
			typ,
			description,
			reserved,
			multipl,
			prometheus.NewDesc(
				prometheus.BuildFQName("", "pingvin", promdesc),
				description,
				nil,
				nil,
			),
		}
	}
	return pingvinRegister{addr, symbol, 0, "0000000000000000", typ, description, reserved, multipl, nil}
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

// Create modbus.Handler, store it in p.handler,
// connect the handler and create p.modbusclient (modbus.Client)
func (p *PingvinKL) createModbusClient() {
	// TODO: read configuration from file, hardcoded for now
	p.handler = modbus.NewRTUClientHandler("/dev/ttyS0")
	p.handler.BaudRate = 19200
	p.handler.DataBits = 8
	p.handler.Parity = "N"
	p.handler.StopBits = 1
	p.handler.SlaveId = 1
	p.handler.Timeout = 1500 * time.Millisecond
	err := p.handler.Connect()
	if err != nil {
		log.Fatal("createModbusClient: p.handler.Connect: ", err)
	}
	p.Debug.Println("Handler connected")
	p.modbusclient = modbus.NewClient(p.handler)
}

func (p *PingvinKL) Quit() {
	err := p.handler.Close()
	if err != nil {
		log.Println("ERROR: Quit:", err)
	}
}

// Update all coil values
func (p *PingvinKL) updateCoils() {
	p.buslock.Lock()
	results, err := p.modbusclient.ReadCoils(0, uint16(len(p.Coils)))
	p.buslock.Unlock()
	if err != nil {
		log.Fatal("updateCoils: client.ReadCoils: ", err)
	}
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

// Read a single holding register, stores value in p.Registers
// Returns integer value of register
func (p *PingvinKL) ReadRegister(addr uint16) (int, error) {
	p.buslock.Lock()
	results, err := p.modbusclient.ReadHoldingRegisters(addr, 1)
	p.buslock.Unlock()
	if err != nil {
		log.Println("ERROR: ReadRegister:", err)
		return 0, err
	}
	if p.Registers[addr].Type == "uint16" {
		p.Registers[addr].Value = int(uint16(results[0]) << 8)
		p.Registers[addr].Value += int(uint16(results[1]))
	} else if p.Registers[addr].Type == "int16" {
		p.Registers[addr].Value = int(int16(results[0]) << 8)
		p.Registers[addr].Value += int(int16(results[1]))
	}
	return p.Registers[addr].Value, nil
}

// Update a single holding register
func (p *PingvinKL) WriteRegister(addr uint16, value uint16) (uint16, error) {
	p.buslock.Lock()
	_, err := p.modbusclient.WriteSingleRegister(addr, value)
	p.buslock.Unlock()
	if err != nil {
		log.Println("ERROR: WriteRegister:", err)
		return 0, err
	}
	val, err := p.ReadRegister(addr)
	if err != nil {
		log.Println("ERROR: WriteRegister:", err)
		return 0, err
	}
	if val == int(value) {
		return value, nil
	}
	return 0, fmt.Errorf("Failed to write register")
}

// Update all holding register values
func (p *PingvinKL) updateRegisters() {
	var err error
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
			p.Debug.Println("Reading registers, attempt", retries, "k:", k)
			p.buslock.Lock()
			results, err = p.modbusclient.ReadHoldingRegisters(uint16(k), uint16(r))
			p.buslock.Unlock()
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
}

// Wrapper function for updating coils, registers and populating
// p.Status for Home Assistant
func (p *PingvinKL) Update() {
	p.updateCoils()
	p.updateRegisters()
	p.populateStatus()
}

// Read single coil
func (p PingvinKL) ReadCoil(n uint16) ([]byte, error) {
	p.buslock.Lock()
	results, err := p.modbusclient.ReadCoils(n, 1)
	p.buslock.Unlock()
	if err != nil {
		log.Fatal("ReadCoil: client.ReadCoils: ", err)
		return nil, err
	}
	p.Coils[n].Value = results[0] == 1
	return results, nil
}

// Force a single coil
func (p *PingvinKL) WriteCoil(n uint16, val bool) bool {
	if val {
		p.checkMutexCoils(n, p.handler)
	}
	var value uint16 = 0
	if val {
		value = 0xff00
	}
	p.buslock.Lock()
	results, err := p.modbusclient.WriteSingleCoil(n, value)
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

// Force multiple coils
func (p *PingvinKL) WriteCoils(startaddr uint16, quantity uint16, vals []bool) error {
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
		p.Debug.Println("index:", i/8, "value:", bits[i/8], "shift:", i%8)
	}
	p.Debug.Println(bits)
	p.buslock.Lock()
	results, err := p.modbusclient.WriteMultipleCoils(startaddr, quantity, bits)
	p.buslock.Unlock()
	if err != nil {
		log.Println("ERROR: WriteCoils: ", err)
		return err
	}
	log.Println(results)
	return nil
}

// Some of the coils are mutually exclusive, and can only be 1 one at a time.
// Check if coil is one of them and force all of them to 0 if so
func (p *PingvinKL) checkMutexCoils(addr uint16, handler *modbus.RTUClientHandler) error {
	for _, mutexcoil := range mutexcoils {
		if mutexcoil == addr {
			for _, n := range mutexcoils {
				if p.Coils[n].Value {
					p.buslock.Lock()
					_, err := p.modbusclient.WriteSingleCoil(n, 0)
					p.buslock.Unlock()
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

// populate p.Status struct for Home Assistant
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
	p.Status.Measurements.SupplyHum = float32(p.Registers[36].Value) / float32(p.Registers[36].Multiplier)
	p.Status.Measurements.Watertemp = float32(p.Registers[12].Value) / float32(p.Registers[12].Multiplier)
	p.Status.Measurements.ExtractIntake = float32(p.Registers[10].Value) / float32(p.Registers[10].Multiplier)
	p.Status.Measurements.ExtractHrc = float32(p.Registers[9].Value) / float32(p.Registers[9].Multiplier)
	p.Status.Measurements.ExtractHum = float32(p.Registers[13].Value) / float32(p.Registers[13].Multiplier)
	p.Status.Measurements.ExtractHum48h = float32(p.Registers[35].Value) / float32(p.Registers[35].Multiplier)
	p.Status.HrcEffIn = p.Registers[29].Value / p.Registers[29].Multiplier
	p.Status.HrcEffEx = p.Registers[30].Value / p.Registers[30].Multiplier
	p.Status.OpMode = parseStatus(p.Registers[44].Value)
	// TODO: Alarms, n of alarms
	// TODO: Uptime & date in separate functions
	p.Status.Coils = p.Coils
}

// Parse readable status from integer (bitfield) value
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

// Change temperature setpoint (register 135)
// action can be up, down or a value.
// If value, the value can be the raw register value (200-300),
// a decimal degree value (20.0 - 23.0), or full degrees (20-30)
// Temperature must be between 20 and 30 deg Celsius, otherwise
// returns an error
func (p *PingvinKL) Temperature(action string) error {
	temperature := 0
	if action == "up" {
		temperature = p.Registers[135].Value + 1*p.Registers[135].Multiplier
		p.Debug.Println("Raising temperature to", temperature)
	} else if action == "down" {
		temperature = p.Registers[135].Value - 1*p.Registers[135].Multiplier
		p.Debug.Println("Lowering temperature to", temperature)
	} else {
		t, err := strconv.Atoi(action)
		if err != nil {
			p.Debug.Println(err)
			tfloat, err := strconv.ParseFloat(action, 32)
			if err != nil {
				p.Debug.Println(err)
				return err
			}
			t = int(tfloat * float64(p.Registers[135].Multiplier))
		}
		if t <= 30 && t >= 20 {
			temperature = 10 * t
		} else {
			temperature = t
		}
		p.Debug.Println("Setting temperature to", temperature)
	}
	if temperature > 300 || temperature < 200 {
		return fmt.Errorf("Temperature setpoint must be between 200 and 300")
	}
	p.Debug.Println("Writing register 135 to", temperature)
	res, err := p.WriteRegister(135, uint16(temperature))
	if err != nil {
		return err
	}
	p.Debug.Println("Temperature changed to", res)
	return nil
}

func (p *PingvinKL) Monitor(interval int) {
	for {
		time.Sleep(time.Duration(interval) * time.Second)
		p.Debug.Println("Updating values")
		p.Update()
	}
}

// Implements prometheus.Describe()
func (p *PingvinKL) Describe(ch chan<- *prometheus.Desc) {
	for _, hreg := range p.Registers {
		if !hreg.Reserved {
			ch <- hreg.PromDesc
		}
	}
	for _, coil := range p.Coils {
		if !coil.Reserved {
			ch <- coil.PromDesc
		}
	}
}

// Implements prometheus.Collect()
func (p *PingvinKL) Collect(ch chan<- prometheus.Metric) {
	for _, hreg := range p.Registers {
		if !hreg.Reserved {
			ch <- prometheus.MustNewConstMetric(
				hreg.PromDesc,
				prometheus.GaugeValue,
				float64(hreg.Value)/float64(hreg.Multiplier),
			)
		}
	}
	for _, coil := range p.Coils {
		val := 0
		if coil.Value {
			val = 1
		}
		if !coil.Reserved {
			ch <- prometheus.MustNewConstMetric(
				coil.PromDesc,
				prometheus.GaugeValue,
				float64(val),
			)
		}
	}
}

// create a PingvinKL struct, read coils and registers from CSVs
func New(debug bool) PingvinKL {
	pingvin := PingvinKL{}
	pingvin.Debug.dbg = debug
	pingvin.buslock = &sync.Mutex{}
	pingvin.createModbusClient()
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
