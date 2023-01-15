import minimalmodbus
import logging
from flask import jsonify

class PingvinCoil():
    """Single coil data structure"""
    def __init__(self, symbol="reserved", description="reserved"):
        self.symbol = symbol
        self.value = False
        self.description = description
        self.reserved = symbol == "reserved" and description == "reserved"

    def serialize(self):
        return {
                    "value": self.value,
                    "symbol": self.symbol,
                    "description": self.description,
                    "reserved": self.reserved
                }

    def get(self):
         return jsonify(self.serialize())

    def flip(self):
        self.value = not self.value

class PingvinCoils():
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
        PingvinCoil("COIL_STOP", "Stop"),
        PingvinCoil("COIL_AWAY", "Away mode"),
        PingvinCoil("COIL_AWAY_L", "Away Long mode"),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil("COIL_MAX_H", "Max Heating"),
        PingvinCoil("COIL_MAX_C", "Max Cooling"),
        PingvinCoil("COIL_CO_BOOST_EN", "CO2 boost"),
        PingvinCoil("COIL_RH_BOOST_EN", "Relative humidity boost"),
        PingvinCoil("COIL_M_BOOST", "Manual boost 100%"),
        PingvinCoil("COIL_TEMP_BOOST_EN", "Temperature boost"),
        PingvinCoil("COIL_SNC", "Summer night cooling"),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil("COIL_AWAY_H", "Heating enabled/disabled in AWAY mode"),
        PingvinCoil("COIL_AWAY_C", "Cooling enabled/disabled in AWAY mode"),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil("COIL_LTO_ON", "Heat recycler state (running=1, stopped = 0)"),
        PingvinCoil(),
        PingvinCoil("COIL_HEAT_ON", "After heater element state (On = 1, Off = 0)"),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil("COIL_TEMP_DECREASE", "Temperature decrease function"),
        PingvinCoil("COIL_OVERTIME", "Programmatic equivalent of OVERTIME digital input"),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil("COIL_ECO_MODE", "Eco mode"),
        PingvinCoil("COIL_ALARM_A", "Alarm of class A active"),
        PingvinCoil("COIL_ALARM_B", "Alarm of class B active"),
        PingvinCoil("COIL_CLK_PROG", "Clock program is currently active"),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil("COIL_SILENT_MODE", "Silent mode"),
        PingvinCoil("COIL_STOP_SLP_COOLING", "Electrical heater cool-off function enabled when the machine has stopped"),
        PingvinCoil("COIL_SERVICE_EN", "Service reminder"),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil("COIL_COOLING_EN", "Active cooling function enabled"),
        PingvinCoil("COIL_LTO_EN"),
        PingvinCoil("COIL_HEATING_EN", "Active heating function enabled"),
        PingvinCoil("COIL_LTO_DEFROST_EN", "HRC defrosting function enabled during winter season"),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil(),
        PingvinCoil()
    ]

    def __init__(self, device, debug=False):
        self.pingvin = device
        if debug: self.coillogger.debug("Updating coil values from device")
        self.update(debug)

    def __getitem__(self, item):
        return self.coils[item]

    def update(self, debug=False):
        """Fetch all coils values from device"""
        self.pingvin.serial.timeout = 0.2
        self.pingvin.debug = debug
        if debug: self.coillogger.info(f"{len(self.coils)} coils registered")
        curvalues = self.pingvin.read_bits(0,len(self.coils),1)
        for i, coil in enumerate(self.coils):
            self.coils[i].value = bool(curvalues[i])
        if debug: self.coillogger.info("Coil values read succesfully")

    def fetchValue(self, address, debug=False):
        """Update single coil value from device and return it"""
        self.pingvin.debug = debug
        if debug: self.coillogger.debug("Updating coil value from device to cache")
        self.coils[address].value = bool(self.pingvin.read_bit(address, 1))
        return self.value(address, debug)

    def value(self, address, debug=False):
        """Get single local coil value"""
        if debug: self.coillogger.debug("Reading coil value from cache")
        return self.coils[address].value

    def print(self, debug=False):
        """Human-readable print of all coil values"""
        coilvals = ""
        for i, coil in enumerate(self.coils):
            coilvals = coilvals + f"Coil {i}\t{coil.value} [{coil.symbol}] ({coil.description})\n"
        return coilvals

    def serialize(self, include_reserved=False):
        """Returns coil values as parseable Python object"""
        coilvals = []
        for i, coil in enumerate(self.coils):
            if not coil.reserved or include_reserved:
                coil = coil.serialize()
                coil['address'] = i
                coilvals.append(coil)
        return coilvals

    def get(self, include_reserved=False, live=False, debug=False):
        """Return all coil values in JSON format"""
        if live: self.update(debug)
        return jsonify(self.serialize(include_reserved))

    def write(self, address):
        self.pingvin.write_bit(address, int(not self.coils[address].value))
        if self.pingvin.read_bit(address, 1) != self.coils[address].value:
            self.coils[address].flip()
            return True
        return False

class PingvinKL():
    """Class for communicating with an Enervent Pinvin Kotilämpö ventilation/heating unit"""
    def __init__(self, serialdevice='/dev/ttyS0', modbusaddr=1, debug=False):
        self.pingvin = minimalmodbus.Instrument(serialdevice, modbusaddr)
        self.coils = PingvinCoils(self.pingvin, debug)