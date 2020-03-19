# Arduino pluggabe discovery for serial ports

The `serial-discovery` tool is a command line program that interacts via stdio. It accepts commands as plain ASCII strings terminated with LF `\n` and sends response as JSON.

## How to build

Install a recent go enviroment (>=13.0) and run `go build`. The executable `serial-discovery` will be produced in your working directory.

## Usage

After startup, the tool waits for commands. The available commands are: `START`, `STOP`,  `QUIT`, `LIST` and `START_SYNC`.

#### START command

The `START` starts the internal subroutines of the discovery that looks for ports. This command must be called before `LIST` or `START_SYNC`. The response to the start command is:

```json
{
  "eventType": "start",
  "message": "OK"
}
```

#### STOP command

The `STOP` command stops the discovery internal subroutines and free some resources. This command should be called if the client wants to pause the discovery for a while. The response to the stop command is:

```json
{
  "eventType": "stop",
  "message": "OK"
}
```

#### QUIT command

The `QUIT` command terminates the discovery. The response to quit is:

```json
{
  "eventType": "quit",
  "message": "OK"
}
```

after this output the tool quits.

#### LIST command

The `LIST` command returns a list of the currently available serial ports. The format of the response is the following:

```json
{
  "eventType": "list",
  "ports": [
    {
      "address": "/dev/ttyACM0",
      "label": "/dev/ttyACM0",
      "prefs": {
        "productId": "0x804e",
        "serialNumber": "EBEABFD6514D32364E202020FF10181E",
        "vendorId": "0x2341"
      },
      "identificationPrefs": {
        "pid": "0x804e",
        "vid": "0x2341"
      },
      "protocol": "serial",
      "protocolLabel": "Serial Port (USB)"
    }
  ]
}
```

The `ports` field contains a list of the available serial ports. If the serial port comes from an USB serial converter the USB VID/PID and USB SERIAL NUMBER properties are also reported inside `prefs`. Inside the `identificationPrefs` instead we have only the properties useful for product identification (in this case USB VID/PID only that may be useful to identify the board)

The list command is a one-shot command, if you need continuos monitoring of ports you should use `START_SYNC` command.

#### START_SYNC command

The `START_SYNC` command puts the tool in "events" mode: the discovery will send `add` and `remove` events each time a new port is detected or removed respectively.

The `add` events looks like the following:

```json
{
  "eventType": "add",
  "port": {
    "address": "/dev/ttyACM0",
    "label": "/dev/ttyACM0",
    "prefs": {
      "productId": "0x804e",
      "serialNumber": "EBEABFD6514D32364E202020FF10181E",
      "vendorId": "0x2341"
    },
    "identificationPrefs": {
      "pid": "0x804e",
      "vid": "0x2341"
    },
    "protocol": "serial",
    "protocolLabel": "Serial Port (USB)"
  }
}
```

it basically gather the same information as the `list` event but for a single port. After calling `START_SYNC` a bunch of `add` events may be generated in sequence to report all the ports available at the moment of the start.

The `remove` event looks like this:

```json
{
  "eventType": "remove",
  "port": {
    "address": "/dev/ttyACM0"
  }
}
```

in this case only the `address` field is reported.

### Example of usage

A possible transcript of the discovery usage:

```json
$ ./serial-discovery 
START
{
  "eventType": "start",
  "message": "OK"
}
START_SYNC
{
  "eventType": "add",
  "port": {
    "address": "/dev/ttyACM0",
    "label": "/dev/ttyACM0",
    "prefs": {
      "productId": "0x804e",
      "serialNumber": "EBEABFD6514D32364E202020FF10181E",
      "vendorId": "0x2341"
    },
    "identificationPrefs": {
      "pid": "0x804e",
      "vid": "0x2341"
    },
    "protocol": "serial",
    "protocolLabel": "Serial Port (USB)"
  }
}
{                                  <--- the board has been disconnected here
  "eventType": "remove",
  "port": {
    "address": "/dev/ttyACM0"
  }
}
{                                  <--- the board has been connected again
  "eventType": "add",
  "port": {
    "address": "/dev/ttyACM0",
    "label": "/dev/ttyACM0",
    "prefs": {
      "productId": "0x804e",
      "serialNumber": "EBEABFD6514D32364E202020FF10181E",
      "vendorId": "0x2341"
    },
    "identificationPrefs": {
      "pid": "0x804e",
      "vid": "0x2341"
    },
    "protocol": "serial",
    "protocolLabel": "Serial Port (USB)"
  }
}
QUIT
{
  "eventType": "quit",
  "message": "OK"
}
$
```

## License

Copyright (c) 2018 ARDUINO SA (www.arduino.cc)

The software is released under the GNU General Public License, which covers the main body
of the serial-discovery code. The terms of this license can be found at:
https://www.gnu.org/licenses/gpl-3.0.en.html

See [LICENSE.txt](<https://github.com/arduino/serial-discovery/blob/master/LICENSE.txt>) for details.

