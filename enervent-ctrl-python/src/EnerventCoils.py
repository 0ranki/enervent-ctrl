import minimalmodbus
import logging

class EnerventCoil():
    """Single coil data structure"""
    def __init__(self, symbol="reserved", description="reserved"):
        self.symbol = symbol
        self.value = 0
        self.description = description
        self.reserved = symbol == "reserved" and description == "reserved"
class Coils():
    """Class for handling Modbus coils"""
    coillogger = logging.getLogger(__name__)
    logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s %(message)s',
    datefmt='%y/%m/%d %H:%M:%S'
    )
    ## coil descriptions and symbols courtesy of Ensto Enervent
    ## https://doc.enervent.com/out/out.ViewDocument.php?documentid=59
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
        if debug: self.coillogger.debug("Updating coil values from device")
        self.update(debug)

    def update(self, debug=False):
        """Fetch all coils values from device"""
        self.pingvin.serial.timeout = 0.2
        self.pingvin.debug = debug
        if debug: self.coillogger.info(f"{len(self.coils)} coils registered")
        curvalues = self.pingvin.read_bits(0,len(self.coils),1)
        for i, coil in enumerate(self.coils):
            self.coils[i].value = curvalues[i]
        self.coillogger.info("Coil values read succesfully")

    def fetchValue(self, address, debug=False):
        """Update single coil value from device and return it"""
        self.pingvin.debug = debug
        if debug: self.coillogger.debug("Updating coil value from device to cache")
        self.coils[address].value = self.pingvin.read_bit(address, 1)
        return self.value(address, debug)

    def value(self, address, debug=False):
        """Return local coil value"""
        if debug: self.coillogger.debug("Reading coil value from cache")
        return self.coils[address].value

    def print(self, debug=False):
        """Human-readable print of all coil values"""
        for i, coil in enumerate(self.coils):
            print(f"Coil {i}\t{coil.value} [{coil.symbol}] ({coil.description})")

class PingvinKL():
    """Class for communicating with an Enervent Pinvin Kotilämpö ventilation/heating unit"""
    def __init__(self, serialdevice='/dev/ttyS0', modbusaddr=1, debug=False):
        self.coils = Coils(serialdevice, modbusaddr, debug)