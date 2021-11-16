package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
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
	StatusMap  map[string]string `yaml:"statusMap"`
	CommandMap map[string]string
}

func (s *server) readConfig() error {
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return fmt.Errorf("reading config.yaml: %w", err)
	}
	if err := yaml.UnmarshalStrict(data, &s.config); err != nil {
		return fmt.Errorf("parsing config.yaml: %w", err)
	}
	fmt.Printf("%+v\n", s.config)
	for _, r := range s.config.Rooms {
		for _, b := range r.Blocks {
			for _, d := range b.Devices {
				if d.Status != "" {
					s.prepareStatus(d.Status)
				}
			}
		}
	}
	return nil
}
