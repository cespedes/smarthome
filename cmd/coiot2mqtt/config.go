package main

import (
	"flag"
	"log"
	"strings"
)

type config struct {
	Debug      bool
	MQTTServer string
}

func readConfig() *config {
	var c config
	flag.BoolVar(&c.Debug, "debug", false, "Debugging")
	flag.StringVar(&c.MQTTServer, "mqtt", "", "MQTT server and prefix (mqtt://host:port/prefix)")
	flag.Parse()

	if c.Debug {
		log.Printf("config = %+v\n", c)
	}
	if c.MQTTServer == "" {
		log.Fatal("No MQTT server specified")
	}

	if !strings.Contains(c.MQTTServer, ":") {
		c.MQTTServer += ":1883"
	}
	if !strings.Contains(c.MQTTServer, "://") {
		c.MQTTServer = "mqtt://" + c.MQTTServer
	}

	return &c
}
