package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/cespedes/smarthome"
)

func main() {
	log.Println("Reading config...")
	c := readConfig()

	fmt.Printf("c = %+v\n", c)

	type point struct {
		nameTag string
		field   string
	}

	topics := make(map[string]point)

	for nameTag, rest := range c.Series {
		for key, value := range rest {
			topics[value] = point{nameTag: nameTag, field: key}
		}
	}
	fmt.Printf("topics = %+v\n", topics)

	log.Println("Creating MQTT client...")
	mqtt, err := smarthome.NewMQTTClient(c.MQTT.Addr, c.MQTT.Root)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Creating Influx client...")
	influx, err := smarthome.NewInfluxClient(c.Influx.Addr, c.Influx.User, c.Influx.Pass, c.Influx.Database)
	if err != nil {
		log.Panic(err)
	}

	var sub string
	if c.MQTT.Root == "" {
		sub = "#"
	} else {
		sub = fmt.Sprintf("%s/#", c.MQTT.Root)
	}
	log.Printf("Subscribing to %q.", sub)
	ch := mqtt.Subscribe(sub)
	for m := range ch {
		myTopic := strings.TrimPrefix(m.Topic, fmt.Sprintf("%s/", c.MQTT.Root))
		if t, ok := topics[myTopic]; ok {
			v := number(string(m.Payload))
			log.Printf("Writing to Influx: (%s %s=%v (%T))", t.nameTag, t.field, v, v)
			go func() {
				err := influx.Insert(t.nameTag, map[string]interface{}{t.field: v})
				if err != nil {
					log.Printf("Writing to influx (%s %s=%v): %s", t.nameTag, t.field, v, err.Error())
				}
			}()
		}
	}
}

func number(v string) interface{} {
	i, err := strconv.Atoi(v)
	if err == nil {
		return i
	}
	f, err := strconv.ParseFloat(v, 64)
	if err == nil {
		return f
	}
	return v
}
