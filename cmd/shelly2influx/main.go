package main

import (
	"log"
	"time"

	"github.com/cespedes/smarthome"
)

func main() {
	log.Println("Reading config...")
	c := readConfig()

	log.Println("Creating Influx client...")
	influx, err := smarthome.NewInfluxClient(c.Influx.Addr, c.Influx.User, c.Influx.Pass, c.Influx.Database)
	if err != nil {
		log.Panic(err)
	}
	// fmt.Printf("influx=%+v\n", influx)

	tick := time.Tick(5 * time.Second)
	for range tick {
		log.Println("tick")
		for _, s := range c.Shelly {
			shelly, err := smarthome.ShellyGetInfo(s.Host)
			if err != nil {
				log.Println(err)
				continue
			}
			// fmt.Printf("got info from %s: %v\n", s.Host, shelly)
			if s.Emeter0 != "" {
				if len(shelly.Status.Emeters) < 1 {
					log.Printf("Shelly device %s does not have Emeter0\n", s.Host)
					continue
				}
				log.Printf("%s: power=%.02f,voltage=%.02f,energy=%.02f\n", s.Emeter0, shelly.Status.Emeters[0].Power, shelly.Status.Emeters[0].Voltage, shelly.Status.Emeters[0].Total)
				err := influx.Insert(s.Emeter0, map[string]interface{}{
					"power":   shelly.Status.Emeters[0].Power,
					"voltage": shelly.Status.Emeters[0].Voltage,
					"energy":  shelly.Status.Emeters[0].Total,
				})
				if err != nil {
					log.Println(err)
				}
			}
			if s.Emeter1 != "" {
				if len(shelly.Status.Emeters) < 2 {
					log.Printf("Shelly device %s does not have Emeter1\n", s.Host)
					continue
				}
				log.Printf("%s: power=%.02f,voltage=%.02f,energy=%.02f\n", s.Emeter1, shelly.Status.Emeters[1].Power, shelly.Status.Emeters[1].Voltage, shelly.Status.Emeters[1].Total)
				err := influx.Insert(s.Emeter1, map[string]interface{}{
					"power":   shelly.Status.Emeters[1].Power,
					"voltage": shelly.Status.Emeters[1].Voltage,
					"energy":  shelly.Status.Emeters[1].Total,
				})
				if err != nil {
					log.Println(err)
				}
			}
		}
		for len(tick) > 0 {
			<-tick
		}
	}
}
