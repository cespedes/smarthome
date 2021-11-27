package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/cespedes/smarthome"
)

type server struct {
	config typeConfig
}

// parseValue tries to parse s as JSON and returns it.
// If it is not a valid JSON, it returns s.
func parseValue(s string) interface{} {
	var parsed interface{}
	err := json.Unmarshal([]byte(s), &parsed)
	if err != nil {
		return s
	}
	return parsed
}

func writeLog(message string, value string) {
	if strings.Contains(message, "{{") {
		tmpl, err := template.New("").Parse(message)
		if err != nil {
			log.Printf("error parsing template %q: %s", message, err.Error())
			return
		}
		var b bytes.Buffer
		parsed := parseValue(value)
		err = tmpl.Execute(&b, parsed)
		if err != nil {
			log.Printf("error executing template %q with value %v: %s", message, parsed, err.Error())
			return
		}
		message = b.String()
	}
	log.Printf("Log: %s\n", message)
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
				writeLog(message, value)
			} else if message, ok := table["."]; ok {
				writeLog(message, value)
			}
		}
	}
}
