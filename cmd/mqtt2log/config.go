package main

import (
	"fmt"
	"log"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type typeConfig struct {
	MQTT struct {
		Server string
		Port   int
	}
	Topics map[string]map[string]string
}

func (s *server) readConfig() error {
	log.Println("Reading config file")
	s.config = typeConfig{}
	data, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return fmt.Errorf("reading config.yaml: %w", err)
	}
	if err := yaml.UnmarshalStrict(data, &s.config); err != nil {
		return fmt.Errorf("parsing config.yaml: %w", err)
	}
	return nil
}
