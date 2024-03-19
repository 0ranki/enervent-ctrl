# Enervent-control

### Upstream URL: [https://git.oranki.net/jarno/enervent-ctrl](https://git.oranki.net/jarno/enervent-ctrl)

External control of an Enervent Pingvin
Kotilämpö residential heating/ventilation
unit via RS485 bus using the Modbus protocol.

Provides a REST API for integration into Home Assistant,
with measurements and basic control over Pingvin functions.

Template YAML configurations for Home Assistant are included
in the `homeassistant` folder, intended to be simple to copy-paste
into Home Assistant's `configuration.yaml` with minimal necessary
modifications. These include sensor configurations, helpers and automations for button functions
and a ready made basic dashboard. No custom components are necessary.

![image](https://user-images.githubusercontent.com/50285623/228834067-503f9820-292c-4614-9316-6cec683e89ef.png)

The daemon is designed to run on a Linux host
that has some sort of RS-485 connector attached.
For development a RPi Zero W 1 with a
connected [Zihatec RS 485 HAT](https://www.hwhardsoft.de/english/projects/rs485-shield/?mobile=1)
has been used.

### Building
- clone or download the repo
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
  -disable-auth
    	Disable HTTP basic authentication (default true)
  -enable-metrics
    	Enable the built-in Prometheus exporter (default true)
  -httplog
    	Enable HTTP access logging
  -interval int
    	Set the interval of background updates (default 4)
  -key string
    	Path to SSL private key to use for HTTPS (default "~/.config/enervent-ctrl/privatekey.pem")
  -logfile string
    	Path to log file. Default is empty string, log to stdout
  -password string
    	Password for HTTP Basic Authentication (default "enervent")
  -read-only
    	Read only mode, no writes to device are allowed
  -regenerate-certs ~/.config/enervent-ctrl/server.crt
    	Generate a new SSL certificate. A new one is generated on startup as ~/.config/enervent-ctrl/server.crt if it doesn't exist.
  -serial string
    	Path to serial console for RS-485 connection. Defaults to /dev/ttyS0 (default "/dev/ttyS0")
  -username string
    	Username for HTTP Basic Authentication (default "pingvin")
```
On first run, the daemon generates `~/.config/enervent-ctrl/configuration.yaml` with default values.
Configuration options are the same as with CLI flags. CLI flags take precedence over the config file.
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

### Running
- Upload the built executable along with `coils.csv` and `registers.csv` to the target host. The files should
  be placed in the same folder.
- Run the binary as a regular user. Adding the user to the correct group for serial access may be necessary
- To run persistently, you can use `screen`, `tmux`, or generate a user systemd service unit file.
- Example systemd service file (named e.g. enervent-ctrl.service):
```
[Unit]
Description=Enervent-ctrl
After=network-online.target

[Service]
Type=simple
Restart=on-failure
RestartSec=30
ExecStart=/path/to/enervent-ctrl-executable

[Install]
WantedBy=default.target
```
- Replace paths in the file and place it under `~/.config/systemd/user`. Create the folder if it doesn't exist.
- `systemctl --user daemon-reload`
- `systemctl --user enable --now enervent-ctrl.service`
- To let user services continue running after logging out:
  - `sudo loginctl enable-linger $USER`

***
# Disclaimer:

**I am not responsible of possible damage to your device if you choose to follow these instructions**

**The manufacturer may void your warranty if you choose to follow these instructions**
***

### Connecting to the Pingvin unit
#### RPi/computer running the daemon
- Connect an RS-485 adapter to the computer you intend to run the daemon on
  - Tested on:
    - RPi 4B and Zero W 1, generic x86_64 linux machines (Alma Linux 8 & 9, Fedora)
    - Zihatec RS-485 HAT with the Pis
    - generic USB-RS485 adapter (checksum errors considerably more often, but nothing critical)
- Ensure the user you intend to run the daemon as has read/write privileges to the serial device.
  - **Not recommended and no need to run as root**
  - Usually adding the user running the executable to the `dialout` group gives permissions to serial devices

#### Pingvin
- Shut down the main power of the unit
- Disconnect the device from mains, discharge any static electricity before proceeding
  - A new motherboard seems to cost close to 1000€ + labour
- Open the cover in which the power switch is attached to. No need to disconnect the switch, there
should be enough length in the wires to move the lid with the switch connected out of the way
![IMG_20230114_133625](https://user-images.githubusercontent.com/50285623/229897490-33d917be-9dea-4b74-bfed-c7b25f9f45f6.jpg)
- Locate the green RS-485 connector on the motherboard, should be on the right edge
  - Schematics available from Enervent at [https://doc.enervent.com/op/op.ViewOnline.php?documentid=940&version=1](https://doc.enervent.com/op/op.ViewOnline.php?documentid=940&version=1), page 38 (finnish)
![IMG_20230114_133824](https://user-images.githubusercontent.com/50285623/229898136-ce7dc020-6c33-4605-86ff-5285000cbbd2.jpg)
- There should be available outlet holes to pass the wires through on the top of the electronics compartment.
- The connector has a detachable plug part. Grab the top of the connector (the part with the screws) with plyers and carefully pull it out. This will make attaching the wire much easier
- Attach wires by tightening the screws in the connector
- Connect **A connector to A connector and B to B**. (they are not Tx/Rx like in many other serials)
  - **NOTE:** After reading quite a few forum posts, many RS-485 adapters seem to have printed the A and B the wrong way, I wouldn't be surprised if this was the case with Pingvin too.
![IMG_20230114_133936](https://user-images.githubusercontent.com/50285623/229900176-5bac0027-80c6-4702-ab74-0ff2b9739507.jpg)
- Plug the plug back to the Pingvin motherboard and close the cover and screws
![IMG_20230114_135258](https://user-images.githubusercontent.com/50285623/229899975-45126a64-7344-4ca0-bfba-c4e524ebe2f8.jpg)
- Reconnect mains and switch both devices on
- Mixing A and B should be safe and won't break anything, but the daemon won't work. If that's the case, disconnect power again and switch the wires on the RPi end.

### Home Assistant

- There are so many variations for HASS configs, that definite instructions are hard to do.
- All the YAMLs are intended to be copy-pasted to `configuration.yaml` (or files included to configuration.yaml)
- Change the IP address, port, username and password according to your configuration
- Restart Home Assistant (A full reload doesn't seem to be enough for all REST integration features to update)


Work is part of my Bachelor's Thesis at Oulu University
of Applied Sciences.

Pingvin and Kotilämpö are registered trademarks of Enervent Zehnder Oy.
