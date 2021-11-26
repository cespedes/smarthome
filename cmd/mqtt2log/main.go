package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/cespedes/smarthome"
)

type server struct {
	config typeConfig
}

func main() {
	// If the file doesn't exist, create it or append to the file
	logfile, err := os.OpenFile("mqtt.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, logfile))

	log.Println("mqtt2log starting")
	var s server
	if err := s.readConfig(); err != nil {
		log.Fatal(err)
	}
	log.Printf("%+v", s.config)

	log.Println("Creating MQTT client...")
	mqttAddr := fmt.Sprintf("mqtt://%s:%d", s.config.MQTT.Server, s.config.MQTT.Port)
	mqtt, err := smarthome.NewMQTTClient(mqttAddr, "")
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Subscribing to \"#\"")
	ch := mqtt.Subscribe("#")

	oldStatus := make(map[string]string)
	for m := range ch {
		topic := m.Topic
		value := string(m.Payload)
		if value == oldStatus[topic] {
			continue
		}
		oldStatus[topic] = value
		if table, ok := s.config.Topics[topic]; ok {
			if message, ok := table[value]; ok {
				log.Printf("Log: %s\n", message)
			} else if message, ok := table["_"]; ok {
				log.Printf("Log: %s\n", strings.ReplaceAll(message, "_", value))
			}
		}
	}
}
