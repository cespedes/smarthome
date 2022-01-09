package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"text/template"
	"time"

	"github.com/at-wat/mqtt-go"
	"github.com/cespedes/smarthome"
)

type server struct {
	config      typeConfig
	logFile     *os.File
	logFileName string
	mqttClient  *smarthome.MQTTClient
	mqttChan    chan *mqtt.Message
	influx      *smarthome.InfluxClient
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

func toFloat(v interface{}) string {
	if i, ok := v.(int); ok {
		return fmt.Sprintf("%d.0", i)
	}
	if f, ok := v.(float64); ok {
		return fmt.Sprintf("%f", f)
	}
	return "0.0"
}

func tmpl(t string, v string) string {
	value := parseValue(v)
	var funcMap = template.FuncMap{
		"now":   time.Now,
		"float": toFloat,
	}
	tmpl, err := template.New("").Funcs(funcMap).Parse(t)
	if err != nil {
		log.Printf("error parsing template %q: %s", t, err.Error())
		return ""
	}
	var b bytes.Buffer
	err = tmpl.Execute(&b, value)
	if err != nil {
		log.Printf("error executing template %q with value %v: %s", t, v, err.Error())
		return ""
	}
	return b.String()
}

func (s *server) foo() {
	log.Printf("Log filename: %q", tmpl(s.config.Logs.Filename, ""))
	log.Printf("Log prefix: %q", tmpl(s.config.Logs.Prefix, ""))
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

func (s *server) writeLog(message string) {
	filename := tmpl(s.config.Logs.Filename, "")
	prefix := tmpl(s.config.Logs.Prefix, "")
	if s.logFileName != filename {
		var err error
		s.logFile.Close()
		os.MkdirAll(filepath.Dir(filename), 0777)
		s.logFile, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Fatal(err)
		}
		s.logFileName = filename
	}
	fmt.Fprintf(s.logFile, "%s %s\n", prefix, message)
}

func (s *server) influxInit() error {
	log.Println("Creating Influx client...")
	var err error
	s.influx, err = smarthome.NewInfluxClient(s.config.Influx.Addr, s.config.Influx.User, s.config.Influx.Pass, s.config.Influx.Database)
	return err
}

func (s *server) init() error {
	if err := s.readConfig(); err != nil {
		return err
	}
	log.Printf("CONFIG: %+v", s.config)
	s.foo()
	if err := s.mqttInit(); err != nil {
		return err
	}
	if err := s.influxInit(); err != nil {
		return err
	}
	return nil
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

	oldValues := make(map[string]string)
	for {
		select {
		case m := <-s.mqttChan:
			topic := m.Topic
			value := string(m.Payload)
			if conf, ok := s.config.Topics[topic]; ok {
				if conf.Changed != "" {
					old, ok := oldValues[topic]
					if ok && value != old {
						v := tmpl(conf.Changed, value)
						s.writeLog(v)
						log.Printf("LOG: %q", v)
					}
					oldValues[topic] = value
				}
				if conf.Log != "" {
					if value != oldValues[topic] {
						v := tmpl(conf.Log, value)
						s.writeLog(v)
						log.Printf("LOG: %q", v)
						oldValues[topic] = value
					}
				}
				if message, ok := conf.Logs[value]; ok {
					if value != oldValues[topic] {
						v := tmpl(message, value)
						s.writeLog(v)
						log.Printf("LOG: %q", v)
						oldValues[topic] = value
					}
				}
				if conf.Influx != "" {
					log.Printf("INFLUX: %q", tmpl(conf.Influx, value))
					err := s.influx.InsertLine(tmpl(conf.Influx, value))
					if err != nil {
						log.Printf("Error: %s", err.Error())
					}
				}
			}
		case <-signalChan:
			log.Println("SIGHUP received")
			if err := s.init(); err != nil {
				log.Fatal(err)
			}
		}
	}
}
