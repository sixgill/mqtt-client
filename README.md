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

Before proceeding, the `sixgill-ingress-api-key` will need to be appropriately changed in the `demo` script for the first case below or supplied to `mqtt-client` directly for the second case.

From the gateway, either 

`./demo`

or

`./mqtt-client <flags>`

if you need more control.

Do `./demo -help` or `./mqtt-client -help` for help on the flags.

