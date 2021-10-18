package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type config struct {
	Influx struct {
		Addr     string
		User     string
		Pass     string
		Database string
	}
	Shelly []struct {
		Host    string
		Emeter0 string
		Emeter1 string
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
