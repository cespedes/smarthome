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

	fmt.Printf("Type: %s\n", shelly.Settings.Device.Type)
	fmt.Printf("MAC: %s\n", shelly.Settings.Device.MAC)
	if shelly.Settings.Mode != "" {
		fmt.Printf("Mode: %s\n", shelly.Settings.Mode)
	}
	fmt.Printf("SSID: %s\n", shelly.Status.WiFiStatus.SSID)
	fmt.Printf("IP: %s\n", shelly.Status.WiFiStatus.IP)
	fmt.Printf("Signal: %d\n", shelly.Status.WiFiStatus.RSSI)
	if shelly.Status.Temperature != 0.0 {
		fmt.Printf("Temperature: %.02f C\n", shelly.Status.Temperature)
	}
	if shelly.Status.Voltage != 0.0 {
		fmt.Printf("Voltage: %.02f V\n", shelly.Status.Voltage)
	}
	for i, v := range shelly.Status.Inputs {
		fmt.Printf("Input %d:\n", i)
		fmt.Printf("\tInput: %d\n", v.Input)
	}
	for i, v := range shelly.Status.Relays {
		fmt.Printf("Relay %d:\n", i)
		fmt.Printf("\tIsOn: %v\n", v.IsOn)
		fmt.Printf("\tSource: %s\n", v.Source)
		if len(shelly.Status.Meters) > i {
			fmt.Printf("\tPower: %.02f W\n", shelly.Status.Meters[i].Power)
			fmt.Printf("\tTotal: %d Wh\n", shelly.Status.Meters[i].Total)
		}
	}
	for i, v := range shelly.Status.Rollers {
		fmt.Printf("Roller %d:\n", i)
		fmt.Printf("\tState: %v\n", v.State)
		if v.Positioning {
			fmt.Printf("\tPos: %d%%\n", v.CurrentPos)
		}
		if len(shelly.Status.Meters) > 2*i+1 {
			fmt.Printf("\tPower: %.02f W\n", shelly.Status.Meters[2*i].Power+shelly.Status.Meters[2*i+1].Power)
			fmt.Printf("\tTotal: %d Wh\n", shelly.Status.Meters[2*i].Total+shelly.Status.Meters[2*i+1].Total)
		}
	}
	for i, v := range shelly.Status.Emeters {
		fmt.Printf("Emeter %d:\n", i)
		fmt.Printf("\tVoltage: %.02f V\n", v.Voltage)
		fmt.Printf("\tPower: %.02f W\n", v.Power)
		fmt.Printf("\tTotal: %.02f Wh\n", v.Total)
		fmt.Printf("\tTotalReturned: %.02f Wh\n", v.TotalReturned)
	}
	// fmt.Printf("SHELLY: %+v\n", shelly)
}
