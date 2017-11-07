# mqtt-client

## Description

This runs on a gateway, subscribes to certain MQTT topics on localhost, then sends the received data to Sixgill Sense 2.0 via their ingress API.

## Building

- clone repository
- `go get`
- `GOOS=linux go build`
- scp the resulting binary to the target machine along with the `demo` helper script
- modify the sixgill Ingress API key and any other parameters in `demo` if necessary

## Usage

From the gateway, either 

`./demo`

or

`./mqtt-client <flags>`

if you need more control.

Do `./mqtt-client -help` or `./demo -help` for help on the flags.


