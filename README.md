# Enervent-control

External control of an Enervent Pingvin
Kotilämpö residential heating/ventilation
unit via RS485 bus using the Modbus protocol.

Work part of my Bachelor's Thesis at Oulu University
of Applied Sciences.

The Python version under `enervent-ctrl-python`
is an initial proof-of-concept,
mainly to test that the hardware side of things
works as expected. The main daemon is written
in Go and the source is under `enervent-ctrl-go`

The daemon is designed to run on a Linux host
that has some sort of RS485 connector attached.
For development a Raspberry Pi 4B was initially
used for convenience, but after the Go
implementation started, a RPi Zero W 1 with a
connected [Zihatec RS 485 HAT](https://www.hwhardsoft.de/english/projects/rs485-shield/?mobile=1)
has been used to make sure the daemon stays as
lightweight as possible.

Pingvin and Kotilämpö are registered trademarks of Enervent Zehnder Oy.
