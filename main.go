package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/resty.v1"

	"github.com/rs/xid"
	pb "github.com/sixgill/sense-ingress-api/proto"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const jwtFile string = "./jwt"

var jwt string

// default mqtt message handler
var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {

	payload, err := ExtractNodeRedDatum(msg.Payload())
	if err != nil {
		log.Println(err.Error())
	}
	start := time.Now()

	// send payload to sixgill Ingress API server
	statusCode, err := PostEvent(*sixgillIngressAddress+"/v1/iot/events", jwt, payload)
	if err != nil {
		log.Println(err.Error())
	}
	log.Printf("TOPIC: %s MSG: %s STATUSCODE: %d Duration: %s\n", msg.Topic(), payload, statusCode, time.Since(start))

}

var sixgillIngressAddress *string

func main() {

	var err error

	// get flags
	mqttBrokerAddress := flag.String("mqtt-broker-address", "localhost", "IP address of the MQTT broker")
	mqttBrokerPort := flag.String("mqtt-broker-port", "1883", "broker's port")
	mqttTopic := flag.String("mqtt-topic", "", "MQTT topic")
	sixgillIngressAddress = flag.String("sixgill-ingress-address", "", "IP address of the Sixgill Ingress API server")
	sixgillIngressAPIKey := flag.String("sixgill-ingress-api-key", "", "API key for Sixgill Ingress API server")
	forceRegister := flag.Bool("force-register", false, "force registration (to update a bad JWT, for instance)")
	flag.Parse()

	// if no topic specified
	if len(*mqttTopic) == 0 {
		log.Println("no mqtt topic specified (-mqtt-topic)")
		os.Exit(1)
	}

	// if no sixgill-ingress-address specified
	if len(*sixgillIngressAddress) == 0 {
		log.Println("no Sixgill Ingress API server address specified (-sixgill-ingress-address)")
		os.Exit(1)
	}

	// if no sixgill-ingress-api-key specified
	if len(*sixgillIngressAPIKey) == 0 {
		log.Println("no Sixgill Ingress API key specified (-sixgill-ingress-api-key)")
		os.Exit(1)
	}

	// fetch any existing JWT
	jwt, err = GetJwtFromFile()
	if *forceRegister || err != nil {
		log.Println("doing registration (no jwt file present or -force-register specified)")
		// do registration
		url := *sixgillIngressAddress + "/v1/registration"
		statusCode, registrationResponse, err := DoRegistration(url, *sixgillIngressAPIKey)
		if err != nil {
			log.Println("unable to do registration (with error): " + err.Error())
			os.Exit(1)
		}
		if statusCode != 200 {
			log.Println("unable to do registration (with status code): ", statusCode)
			os.Exit(1)
		}

		jwt = registrationResponse.Token

		// save JWT for next time
		err = PutJwtToFile(jwt)
		if err != nil {
			log.Println("unable to write JWT: " + err.Error())
		}

	} else {
		log.Println("got jwt from file")
	}

	log.Println("setting up mqtt broker")
	// set up mqtt broker for specified address and port
	opts := MQTT.NewClientOptions().AddBroker("tcp://" + *mqttBrokerAddress + ":" + *mqttBrokerPort)

	// set a unique mqtt client id (duplicates will stall all parties with that id)
	opts.SetClientID("mqtt-client-" + xid.New().String())
	opts.SetDefaultPublishHandler(f)

	// create and start the client
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Println("unable to do start mqtt client:", token.Error())
		os.Exit(1)
	}

	// subscribe to the mqtt topic
	// TODO: subscribe to multiple topics
	token := c.Subscribe(*mqttTopic, 0, nil)
	token.Wait()
	if token.Error() != nil {
		log.Println("unable to subscribe to topic '"+*mqttTopic+"':", token.Error())
		os.Exit(1)
	}

	// recognize stop-related signals
	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

	// wait / block on those signals
	<-exitSignal

	// indicate cleanup
	log.Println("cleaning up")

	// unsubscribe from mqtt topic
	if token := c.Unsubscribe(*mqttTopic); token.Wait() && token.Error() != nil {
		log.Println("unable to unsubscribe from topic '"+*mqttTopic+"'", token.Error())
		os.Exit(1)
	}

	// some mqtt-related delay
	c.Disconnect(250)

	// free and clear!
	log.Println("bye")
}

// DoRegistration registers this application with the sixgill api server
func DoRegistration(url, apiKey string) (int, pb.RegistrationResponse, error) {

	// curl -X POST "http://sense-ingress-api-staging.sixgill.run/v1/registration"  -d '{
	// 	"apiKey":"01BWHNJHFCZXDVDYPTK8080WC1",
	// 	"properties":{
	// 	  "timestamp":1509038692,
	// 	  "manufacturer":"Apple",
	// 	  "model":"advantech",
	// 	  "os":"wrlinux",
	// 	  "osVersion":"7.0.0.13",
	// 	  "softwareVersion":"sixgill-mqtt-v0.1",
	// 	  "type":"wrlinux",
	// 	  "sensors":["temperature","humidity"]
	//   }
	// }'

	request := &pb.RegistrationRequest{
		ApiKey: apiKey,
		Properties: &pb.Property{
			Timestamp:       int64(time.Now().UTC().Second()),
			Manufacturer:    "Intel",
			Model:           "Advantech",
			Os:              "wrlinux",
			OsVersion:       "7.0.0.13",
			SoftwareVersion: "sixgill-mqtt-v0.1",
			Type:            "wrlinux",
			Sensors:         []string{"temperature", "humidity"},
		},
	}

	response := &pb.RegistrationResponse{}
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(request).
		SetResult(response).
		SetContentLength(true).
		Post(url)

	return resp.StatusCode(), *response, err
}

// GetJwtFromFile gets the previously stored JWT from the file
func GetJwtFromFile() (string, error) {
	jwt, err := ioutil.ReadFile(jwtFile)
	log.Println("read from jwt file")
	return string(jwt), err
}

// PutJwtToFile puts the JWT into the file
func PutJwtToFile(jwt string) error {
	log.Println("writing to jwt file")
	return ioutil.WriteFile(jwtFile, []byte(jwt), 0644)
}

// PostEvent POSTs the event to the ingress API server using the jwt
func PostEvent(url, jwt string, event []byte) (int, error) {

	request := string(event)
	response := &pb.RegistrationResponse{}
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(jwt).
		SetBody(request).
		SetResult(response).
		SetContentLength(true).
		Post(url)

	return resp.StatusCode(), err
}

// ExtractNodeRedDatum adds a `timestamp` and `value` field based on the elements of Node-Red's `datum` field if that field is present
func ExtractNodeRedDatum(payload []byte) ([]byte, error) {

	// extract datum field
	var data map[string]interface{}
	err := json.Unmarshal([]byte(payload), &data)
	if err != nil {
		return payload, errors.New("unable to unmarshal payload")
	}

	datum, present := data["datum"]
	if !present {
		log.Println("no datum present")
		return payload, nil // not an error, just return the payload
	}

	// throw error if `timestamp`` field already exists
	if _, present := data["timestamp"]; present {
		return payload, errors.New("timestamp field already present")
	}

	// throw error if `value`` field already exists
	if _, present := data["value"]; present {
		return payload, errors.New("value field already present")
	}

	// now tease out our values for timestamp ...
	data["timestamp"] = datum.([]interface{})[0]

	// and value
	data["value"] = datum.([]interface{})[1]

	// marshal with added elements
	augmentedPayload, err := json.Marshal(data)
	if err != nil {
		return payload, errors.New("unable to marshal new json")
	}

	return augmentedPayload, nil // FTM
}
