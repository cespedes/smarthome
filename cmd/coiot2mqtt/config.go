package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type config struct {
	MQTT struct {
		Addr string
		Root string
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
