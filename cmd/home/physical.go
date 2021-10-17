package main

import (
	"fmt"
	"net/url"
)

func (s *server) getStatusShelly(url *url.URL) {
	fmt.Printf("will connect to %q and ask for %q\n", url.Host, url.Path)
}

func (s *server) getStatus(rawURL string) {
	url, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	switch url.Scheme {
	case "shelly":
		s.getStatusShelly(url)
	default:
		panic(fmt.Sprintf("Unknown sheme %q", url.Scheme))
	}
}
