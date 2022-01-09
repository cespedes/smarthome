package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type homeConfig struct {
	Auth    map[string]string
	MQTT    interface{} `yaml:"mqtt"`
	Devices []Device
	Rooms   []Room
}

type Device struct {
	ID     string
	Input  string   // MQTT topic where status is published
	Min    int      // Minimum possible value
	Max    int      // Maximum possible value
	Values []string // List of possible values
	Output string   // MQTT topic where we can publish to perform an action
}

type Room struct {
	Name   string
	Blocks []roomBlock
}

type roomBlock struct {
	Name    string
	Devices []blockDevice
}

type blockDevice struct {
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
