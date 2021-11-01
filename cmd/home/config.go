package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type homeConfig struct {
	Auth  map[string]string
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
	ID         string
	Type       string
	Units      string
	Min        int
	Max        int
	Status     string
	Command    string
	StatusMap  map[string]string
	CommandMap map[string]string
}

func (s *server) readConfig() {
	if _, err := toml.DecodeFile("config.toml", &s.config); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%+v\n", s.config)
	for _, r := range s.config.Rooms {
		for _, b := range r.Blocks {
			for _, d := range b.Devices {
				if d.Status != "" {
					s.getStatus(d.Status)
				}
			}
		}
	}
}
