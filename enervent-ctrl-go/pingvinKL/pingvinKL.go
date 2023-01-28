package pingvinKL

import (
	"bufio"
	"log"
	"os"
	"strings"
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

// create a PingvinKL struct, read coils and registers from CSVs
func New() PingvinKL {
	pingvin := PingvinKL{}
	coilData := readCsvLines("coils.csv")
	for i := 0; i < len(coilData); i++ {
		pingvin.Coils = append(pingvin.Coils, newCoil(coilData[i][1], coilData[i][2]))
	}
	return pingvin
}

// func New() PingvinKL {
// 	pingvin := PingvinKL{}
// 	// var Coils []pingvinCoil
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_STOP", "Stop"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_AWAY", "Away mode"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_AWAY_L", "Away Long mode"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_MAX_H", "Max Heating"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_MAX_C", "Max Cooling"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_CO_BOOST_EN", "CO2 boost"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_RH_BOOST_EN", "Relative humidity boost"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_M_BOOST", "Manual boost 100%"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_TEMP_BOOST_EN", "Temperature boost"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_SNC", "Summer night cooling"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_AWAY_H", "Heating enabled/disabled in AWAY mode"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_AWAY_C", "Cooling enabled/disabled in AWAY mode"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_LTO_ON", "Heat recycler state (running=1, stopped = 0)"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_HEAT_ON", "After heater element state (On = 1, Off = 0)"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_TEMP_DECREASE", "Temperature decrease function"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_OVERTIME", "Programmatic equivalent of OVERTIME digital input"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_ECO_MODE", "Eco mode"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_ALARM_A", "Alarm of class A active"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_ALARM_B", "Alarm of class B active"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_CLK_PROG", "Clock program is currently active"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_SILENT_MODE", "Silent mode"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_STOP_SLP_COOLING", "Electrical heater cool-off function enabled when the machine has stopped"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_SERVICE_EN", "Service reminder"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_COOLING_EN", "Active cooling function enabled"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_LTO_EN", "N/A"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_HEATING_EN", "Active heating function enabled"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("COIL_LTO_DEFROST_EN", "HRC defrosting function enabled during winter season"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	pingvin.Coils = append(pingvin.Coils, newCoil("-", "-"))
// 	return pingvin
// }
