package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type chaletConfig struct {
	Rooms []configRoom
}

type configRoom struct {
	Name   string
	Blocks []configBlock
}

type configBlock struct {
	Name    string
	Devices []configDevice
}

type configDevice struct {
	ID    string
	Type  string
	Units string
	Min   int
	Max   int
}

func (s *server) readConfig() {
	if _, err := toml.DecodeFile("config.toml", &s.config); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v\n", s.config)
}
