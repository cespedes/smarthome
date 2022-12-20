package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type config struct {
	Interval int
	MQTT     struct {
		Addr string
	}
	Shelly []struct {
		Host string
	}
}

func readConfig() *config {
	var c config
	if _, err := toml.DecodeFile("config.toml", &c); err != nil {
		fmt.Println(err)
		return nil
	}
	return &c
}
