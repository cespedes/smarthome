package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type config struct {
	Interval int
	Influx   struct {
		Addr     string
		User     string
		Pass     string
		Database string
	}
	MQTT struct {
		Addr string
		Root string
	}
	Series map[string]map[string]string
}

func readConfig() *config {
	var c config

	_, err := toml.DecodeFile("config.toml", &c)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return &c
}
