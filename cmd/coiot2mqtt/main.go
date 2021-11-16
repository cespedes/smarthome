package main

import (
	"log"

	"github.com/cespedes/smarthome"
)

func main() {
	log.Println("Reading config...")
	conf := readConfig()

	log.Println("Creating MQTT client...")
	mqtt, err := smarthome.NewMQTTClient(conf.MQTT.Addr, conf.MQTT.Root)
	if err != nil {
		log.Fatal(err)
	}

	coiot, err := smarthome.CoIoTinit()
	if err != nil {
		panic(err)
	}

	for {
		d := coiot.Read()
		if d.Code[0] == 0 && d.Code[1] == 30 {
			log.Printf("CoIoT packet from %s: %q\n", d.RemoteAddr, d.Payload)
			mqtt.Publish("coiot", string(d.Payload))
		}
	}
}
