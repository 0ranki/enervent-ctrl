import minimalmodbus

class EnerventCoil():
    def __init__(self, symbol="", description="", reserved=True):
        self.symbol = symbol
        self.value = 0
        self.description = description
        self.reserved = reserved

class PingvinCoils():
    coils = [
        EnerventCoil("COIL_STOP", "Stop"),
        EnerventCoil("COIL_AWAY", "Away mode"),
        EnerventCoil("COIL_AWAY_L", "Away Long mode"),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil("COIL_MAX_H", "Max Heating"),
        EnerventCoil("COIL_MAX_C", "Max Cooling"),
        EnerventCoil("COIL_CO_BOOST_EN", "CO2 boost"),
        EnerventCoil("COIL_RH_BOOST_EN", "Relative humidity boost"),
        EnerventCoil("COIL_M_BOOST", "Manual boost 100%"),
        EnerventCoil("COIL_TEMP_BOOST_EN", "Temperature boost"),
        EnerventCoil("COIL_SNC", "Summer night cooling"),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil("COIL_AWAY_H", "Heating enabled/disabled in AWAY mode"),
        EnerventCoil("COIL_AWAY_C", "Cooling enabled/disabled in AWAY mode"),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil("COIL_LTO_ON", "Heat recycler state (running=1, stopped = 0)"),
        EnerventCoil(),
        EnerventCoil("COIL_HEAT_ON", "After heater element state (On = 1, Off = 0)"),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil("COIL_TEMP_DECREASE", "Temperature decrease function"),
        EnerventCoil("COIL_OVERTIME", "Programmatic equivalent of OVERTIME digital input"),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil("COIL_ECO_MODE", "Eco mode"),
        EnerventCoil("COIL_ALARM_A", "Alarm of class A active"),
        EnerventCoil("COIL_ALARM_B", "Alarm of class B active"),
        EnerventCoil("COIL_CLK_PROG", "Clock program is currently active"),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil("COIL_SILENT_MODE", "Silent mode"),
        EnerventCoil("COIL_STOP_SLP_COOLING", "Electrical heater cool-off function enabled when the machine has stopped"),
        EnerventCoil("COIL_SERVICE_EN", "Service reminder"),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil("COIL_COOLING_EN", "Active cooling function enabled"),
        EnerventCoil("COIL_LTO_EN"),
        EnerventCoil("COIL_HEATING_EN", "Active heating function enabled"),
        EnerventCoil("COIL_LTO_DEFROST_EN", "HRC defrosting function enabled during winter season"),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil(),
        EnerventCoil()
    ]

    def __init__(self, serialdevice='/dev/ttyS0', modbusaddr=1, debug=False):
        self.pingvin = minimalmodbus.Instrument(serialdevice, modbusaddr)
        self.pingvin.serial.timeout = 0.2
        self.pingvin.debug = debug
        curvalues = self.pingvin.read_bits(0,71,1)
        for i, coil in enumerate(self.coils):
            self.coils[i].value = curvalues[i]

