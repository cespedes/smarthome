package main

import (
	"log"
	"net/url"
)

func (s *server) getMQTTStatus(url *url.URL) {
	log.Printf("MQTT: will connect to %q and ask for %q\n", url.Host, url.Path)
}

func (s *server) getStatus(rawURL string) {
	url, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	switch url.Scheme {
	case "mqtt":
		s.getMQTTStatus(url)
	default:
		log.Printf("getStatus(): Unknown scheme in %q", url)
	}
}
