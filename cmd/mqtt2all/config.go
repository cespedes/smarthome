package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type typeConfig struct {
	MQTT struct {
		Server string
		Port   int
	}
	Influx struct {
		Addr     string
		User     string
		Pass     string
		Database string
	}
	Postgres struct {
		Connect string
		Schema  string
	}
	Logs struct {
		Filename string
		Prefix   string
	}
	Debug struct {
		Filename string
		Prefix   string
	}
	Topics map[string]map[string]struct {
		Log      string
		Debug    string
		Influx   string
		Postgres string
		Exec     string
	}
}

func (s *server) readConfig() error {
	configFileName := "config.yaml"
	if s.verbose {
		log.Printf("Reading config file %s", configFileName)
	}
	s.config = typeConfig{}
	data, err := ioutil.ReadFile(configFileName)
	if err != nil {
		return fmt.Errorf("reading %s: %w", configFileName, err)
	}
	if err := yaml.UnmarshalStrict(data, &s.config); err != nil {
		return fmt.Errorf("parsing %s: %w", configFileName, err)
	}
	return nil
}
