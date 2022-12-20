package main

import (
	"fmt"
	"log"
	"time"

	"github.com/cespedes/smarthome"
)

func main() {
	log.Println("Reading config...")
	c := readConfig()

	log.Println("Creating MQTT client...")
	mqtt, err := smarthome.NewMQTTClient(c.MQTT.Addr)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Beginning loop.")
	tick := time.Tick(time.Duration(c.Interval) * time.Second)
	for range tick {
		log.Println("tick")
		for _, s := range c.Shelly {
			go func(hostname string) {
				shelly, err := smarthome.ShellyGetInfo(hostname)
				if err != nil {
					log.Println(err)
					return
				}
				// log.Printf("got info from %s: %v\n", s.Host, shelly)
				log.Printf("%s: type=%s inputs=%d relays=%d rollers=%d meters=%d emeters=%d\n",
					hostname, shelly.Settings.Device.Type,
					len(shelly.Status.Inputs), len(shelly.Status.Relays),
					len(shelly.Status.Rollers), len(shelly.Status.Meters),
					len(shelly.Status.Emeters))
				mqtt.Publish(fmt.Sprintf("%s/mac", hostname), shelly.Settings.Device.MAC)
				mqtt.Publish(fmt.Sprintf("%s/type", hostname), shelly.Settings.Device.Type)
				if shelly.Status.Voltage != 0.0 {
					mqtt.Publish(fmt.Sprintf("%s/voltage", hostname), fmt.Sprint(shelly.Status.Voltage))
				}
				if shelly.Status.Temperature != 0.0 {
					mqtt.Publish(fmt.Sprintf("%s/temperature", hostname), fmt.Sprint(shelly.Status.Temperature))
				}
				for i, input := range shelly.Status.Inputs {
					mqtt.Publish(fmt.Sprintf("%s/input%d", hostname, i), fmt.Sprint(input.Input))
				}
				for i, relay := range shelly.Status.Relays {
					mqtt.Publish(fmt.Sprintf("%s/relay%d/is-on", hostname, i), fmt.Sprint(relay.IsOn))
					mqtt.Publish(fmt.Sprintf("%s/relay%d/source", hostname, i), relay.Source)
				}
				for i, roller := range shelly.Status.Rollers {
					mqtt.Publish(fmt.Sprintf("%s/roller%d/state", hostname, i), fmt.Sprint(roller.State))
					mqtt.Publish(fmt.Sprintf("%s/roller%d/source", hostname, i), roller.Source)
					if roller.Positioning {
						mqtt.Publish(fmt.Sprintf("%s/roller%d/pos", hostname, i), fmt.Sprint(roller.CurrentPos))
					}
				}
				for i, emeter := range shelly.Status.Emeters {
					mqtt.Publish(fmt.Sprintf("%s/emeter%d/power", hostname, i), fmt.Sprintf("%.02f", emeter.Power))
					mqtt.Publish(fmt.Sprintf("%s/emeter%d/reactive", hostname, i), fmt.Sprintf("%.02f", emeter.Reactive))
					mqtt.Publish(fmt.Sprintf("%s/emeter%d/voltage", hostname, i), fmt.Sprintf("%.02f", emeter.Voltage))
					mqtt.Publish(fmt.Sprintf("%s/emeter%d/total", hostname, i), fmt.Sprintf("%.02f", emeter.Total))
				}
			}(s.Host)
		}
		for len(tick) > 0 {
			<-tick
		}
	}
}
