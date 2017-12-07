# mqtt-client

## Description

This process runs on a gateway, subscribes to certain MQTT topics on localhost, then sends the received data to Sixgill Sense 2.0 via their ingress API.

## Building

- clone repository
- `go get`
- `GOOS=linux go build`
- scp the resulting binary to the target machine
- modify the sixgill Ingress API key and any other parameters in `mqtt-client-conf.json` if necessary

## Getting

Get the latest release from [here](https://github.com/sixgill/mqtt-client/releases).

## Using

- Change values in `mqtt-client-conf.json` as needed (parameters are described below), 
- Copy the `mqtt-client-conf.json` file to `~/.sense/`, and 
- Run `$ ./mqtt-client <optional flags>`

Do `./mqtt-client -help` for help on the flags.

## Parameters 

| Parameter | Default Value | Meaning |
| --------: | :-----------: | :------ |
| mqtt-broker-address | "localhost" | IP address of the MQTT broker |
| mqtt-broker-port | "1883" | broker's port |
| mqtt-topic | "" | MQTT topic |
| sense-ingress-address | "" | IP address of the Sixgill Sense Ingress API server |
| sense-ingress-api-key | "" | API key for Sixgill Sense Ingress API server |

