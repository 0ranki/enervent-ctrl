#!/usr/bin/env python
import logging
from PingvinKL import PingvinKL
from flask import Flask, request

VERSION = "0.0.1"
DEBUG = True

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
def get_all():
    return pingvin.coils.get(include_reserved=request.args.get('include_reserved'),live=request.args.get('live'),debug=DEBUG)

@app.route('/api/v1/coils/<int:address>', methods=["GET","PUT"])
def coil(address):
    if request.method == 'GET':
        coil = pingvin.coils[address].get()
        return coil
    elif request.method == 'PUT':
        return {"success": pingvin.coils.write(address)}

@app.route('/')
def dump():
    return pingvin.coils.print(debug=DEBUG)

if __name__ == "__main__":
    log.info(f"Starting enervent-logger {VERSION}")
    app.run(host='0.0.0.0',port=8888,debug=DEBUG)
