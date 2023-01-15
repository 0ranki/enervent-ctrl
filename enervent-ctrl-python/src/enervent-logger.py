#!/usr/bin/env python
import minimalmodbus
import logging

VERSION = "0.0.1"

## Logging configuration
log = logging.getLogger(__name__)
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s %(message)s',
    datefmt='%y/%m/%d %H:%M:%S'
    )

if __name__ == "__main__":
    log.info(f"Starting enervent-logger {VERSION}")
    