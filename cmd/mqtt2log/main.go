package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/template"

	"github.com/at-wat/mqtt-go"
	"github.com/cespedes/smarthome"
)

type server struct {
	config     typeConfig
	logFile    *os.File
	mqttClient *smarthome.MQTTClient
	mqttChan   chan *mqtt.Message
}

func (s *server) openLog() error {
	var err error

	if s.logFile != nil {
		s.logFile.Close()
	}
	s.logFile, err = os.OpenFile("mqtt.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	log.SetOutput(io.MultiWriter(os.Stdout, s.logFile))
	return nil
}

func (s *server) mqttInit() error {
	var err error
	log.Println("Creating MQTT client...")
	mqttAddr := fmt.Sprintf("mqtt://%s:%d", s.config.MQTT.Server, s.config.MQTT.Port)
	s.mqttClient, err = smarthome.NewMQTTClient(mqttAddr, "")
	if err != nil {
		return err
	}
	log.Printf("Subscribing to \"#\"")
	s.mqttChan = s.mqttClient.Subscribe("#")
	return nil
}

func (s *server) init() error {
	if err := s.openLog(); err != nil {
		return err
	}
	if err := s.readConfig(); err != nil {
		return err
	}
	log.Printf("CONFIG: %+v", s.config)
	if err := s.mqttInit(); err != nil {
		return err
	}
	return nil
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
	log.Println("mqtt2log starting")

	// SIGHUP handling:
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP)

	var s server
	if err := s.init(); err != nil {
		log.Fatal(err)
	}

	oldStatus := make(map[string]string)
	for {
		select {
		case m := <-s.mqttChan:
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
		case <- signalChan:
			log.Println("SIGHUP received")
			if err := s.openLog(); err != nil {
				log.Println(err)
			}
			if err := s.readConfig(); err != nil {
				log.Println(err)
			}
			if err := s.mqttInit(); err != nil {
				log.Println(err)
			}
		}
	}
}
