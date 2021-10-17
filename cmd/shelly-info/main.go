package main

import (
	"fmt"
	"os"

	"github.com/cespedes/smarthome"
)

const ProgName = "shelly-info"

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <shelly-device>\n", ProgName)
		os.Exit(1)
	}
	shelly, err := smarthome.ShellyGetInfo(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	fmt.Printf("SHELLY: %+v\n", shelly)
}
