package main

import (
	"log"
	"net/http"
	"net/url"
)

func (s *server) getMQTTStatus(url *url.URL) {
	log.Printf("MQTT: will connect to %q and ask for %q\n", url.Host, url.Path)
}

func (s *server) prepareStatus(rawURL string) {
	url, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	switch url.Scheme {
	case "mqtt":
		s.getMQTTStatus(url)
	default:
		log.Printf("prepareStatus(): Unknown scheme in %q", url)
	}
}

func (s *server) getStatus(w http.ResponseWriter, r *http.Request) {
}
