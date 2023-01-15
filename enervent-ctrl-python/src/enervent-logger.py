#!/usr/bin/env python
import logging
from EnerventCoils import PingvinKL

VERSION = "0.0.1"
DEBUG = False

## Logging configuration
log = logging.getLogger(__name__)
logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s %(message)s',
    datefmt='%y/%m/%d %H:%M:%S'
    )

if __name__ == "__main__":
    log.info(f"Starting enervent-logger {VERSION}")
    pingvin = PingvinKL('/dev/ttyS0',1,debug=DEBUG)
    print(pingvin.coils.value(1, debug=DEBUG))
    print(pingvin.coils.fetchValue(1, debug=DEBUG))
    print(pingvin.coils.print())
