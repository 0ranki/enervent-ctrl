# Enervent-control

External control of an Enervent Pingvin
Kotilämpö residential heating/ventilation
unit via RS485 bus using the Modbus protocol.

Provides a REST API for integration into Home Assistant,
with measurements and basic control over Pingvin functions.

Template YAML configurations for Home Assistant are included
in the `homeassistant` folder, intended to be simple to copy-paste
into Home Assistant's `configuration.yaml` with minimal necessary
modifications. These include sensor configurations, helpers and automations for button functions
and a ready made basic dashboard.
![image](https://user-images.githubusercontent.com/50285623/228834067-503f9820-292c-4614-9316-6cec683e89ef.png)

The daemon is designed to run on a Linux host
that has some sort of RS-485 connector attached.
For development a RPi Zero W 1 with a
connected [Zihatec RS 485 HAT](https://www.hwhardsoft.de/english/projects/rs485-shield/?mobile=1)
has been used.

### Building
- clone or download the repo
- `static/html/index.html` is symlinked to `coils` and `registers`
  for development purposes, the symlinks need to be dereferenced before
  building the binary on filesystems that support symlinks
  - Replace symlinks with copies of the files or use e.g. `tar -h`
- Build for the correct architecture, e.g. for Linux 32-bit ARM (Rpi Zero W 1):
  ```
  cd /path/to/repo
  env GOOS=linux GOARCH=arm go build -o BUILD/enervent-ctrl-linux-arm32
  ```

### Configuration:
- CLI flags:
```
  -cert string
        Path to SSL public key to use for HTTPS (default "~/.config/enervent-ctrl/certificate.pem")
  -debug
        Enable debug logging
  -enable-metrics
        Enable the built-in Prometheus exporter (default true)
  -httplog
        Enable HTTP access logging
  -interval int
        Set the interval of background updates (default 4)
  -key string
        Path to SSL private key to use for HTTPS (default "/home/jarno/.config/enervent-ctrl/privatekey.pem")
  -logfile string
        Path to log file. Default is empty string, log to stdout
  -password string
        Password for HTTP Basic Authentication (default "enervent")
  -regenerate-certs ~/.config/enervent-ctrl/server.crt
        Generate a new SSL certificate. A new one is generated on startup as ~/.config/enervent-ctrl/server.crt if it doesn't exist.
  -serial string
        Path to serial console for RS-485 connection. Defaults to /dev/ttyS0 (default "/dev/ttyS0")
  -username string
        Username for HTTP Basic Authentication (default "pingvin")
```
On first run, the daemon generates `~/.config/enervent-ctrl/configuration.yaml` with default values.
Configuration options are the same as with CLI flags. CLI flags take precedenence over the config file.
- `serial_address:` Path to RS-485 serial device
- `port:` TCP port for the REST API to listen on
- `ssl_certificate:` Path to SSL certificate for HTTPS
- `ssl_privatekey:` Path to SSL private key for HTTPS
- `username:` Username for REST API HTTP Basic Auth
- `password:` Password for REST API HTTP Basic Auth
- `interval:` Interval of background updates from Modbus
- `enable_metrics:` Enable the built-in Prometheus exporter
- `log_file:` Path to log file, default logging is to STDOUT
- `log_access:` Enable HTTP Access logging to logfile/STDOUT
- `debug:` Enable debug logging

Readme will be updated in the near future with physical connection instructions.

Work part of my Bachelor's Thesis at Oulu University
of Applied Sciences.

Pingvin and Kotilämpö are registered trademarks of Enervent Zehnder Oy.
