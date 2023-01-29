package pingvinKL

import (
	"bufio"
	"log"
	"os"
	"strings"
	"time"

	"github.com/goburrow/modbus"
)

// single coil data
type pingvinCoil struct {
	Symbol      string
	Value       bool
	Description string
	Reserved    bool
}

func newCoil(symbol string, description string) pingvinCoil {
	reserved := symbol == "-" && description == "-"
	coil := pingvinCoil{symbol, false, description, reserved}
	return coil
}

// unit modbus data
type PingvinKL struct {
	Coils []pingvinCoil
}

// read a CSV file containing data for coils or registers
func readCsvLines(file string) [][]string {
	delim := ","
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

// Configure the modbus client on creation
func (p PingvinKL) getHandler() *modbus.RTUClientHandler {
	// TODO: read configuration from file, hardcoded for now
	handler := modbus.NewRTUClientHandler("/dev/ttyS0")
	handler.BaudRate = 19200
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 200 * time.Millisecond
	return handler
}

func (p PingvinKL) Update() {
	// coildata, err := p.Client.ReadCoils(0, len(p.Coils))
}

func (p PingvinKL) ReadCoil(n uint16) []byte {
	handler := p.getHandler()
	err := handler.Connect()
	if err != nil {
		log.Fatal("ReadCoil1: ", err)
	}
	defer handler.Close()
	client := modbus.NewClient(handler)
	results, err := client.ReadCoils(n, 1)
	if err != nil {
		log.Fatal("ReadCoil2: ", err)
	}
	return results
}

// create a PingvinKL struct, read coils and registers from CSVs
func New() PingvinKL {
	pingvin := PingvinKL{}
	coilData := readCsvLines("coils.csv")
	for i := 0; i < len(coilData); i++ {
		pingvin.Coils = append(pingvin.Coils, newCoil(coilData[i][1], coilData[i][2]))
	}
	return pingvin
}
