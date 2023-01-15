#!/usr/bin/env python
import logging
from EnerventCoils import PingvinKL
from flask import Flask, request

VERSION = "0.0.1"
DEBUG = False

## Logging configuration
log = logging.getLogger(__name__)
logging.basicConfig(
    level=logging.DEBUG,
    format='%(asctime)s %(message)s',
    datefmt='%y/%m/%d %H:%M:%S'
    )

pingvin = PingvinKL('/dev/ttyS0',1,debug=DEBUG)
app = Flask(__name__)

@app.route('/api/v1/coils')
def get():
    return pingvin.coils.get()

@app.route('/api/v1/coils/all')
def get_all():
    return pingvin.coils.get(include_reserved=True)

@app.route('/api/v1/coils/<int:address>', methods=["GET","PUT"])
def get_coil():
    if request.method == 'GET':
        return 


if __name__ == "__main__":
    log.info(f"Starting enervent-logger {VERSION}")
    # print(pingvin.coils.value(1, debug=DEBUG))
    # print(pingvin.coils.fetchValue(1, debug=DEBUG))
    # print(pingvin.coils.print())
    app.run(host='0.0.0.0',port=8888,debug=True)
