package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/cespedes/smarthome"
)

type coiotBlock struct {
	ID   int
	Name string
}
type coiotState struct {
	ID    int
	Block string
	Name  string
	Unit  string
}

type coiotServer struct {
	Hostname string
	Blocks   []coiotBlock
	States   []coiotState
}

func main() {
	log.Println("Reading config...")
	conf := readConfig()

	log.Println("Creating MQTT client...")
	mqtt, err := smarthome.NewMQTTClient(conf.MQTT.Addr, conf.MQTT.Root)
	if err != nil {
		log.Fatal(err)
	}

	coiot, err := smarthome.CoIoTinit()
	if err != nil {
		panic(err)
	}

	servers := make(map[string]*coiotServer)

	for {
		d := coiot.Read()
		if d.Code[0] == 0 && d.Code[1] == 30 {
			// log.Printf("CoIoT packet from %s: %q\n", d.RemoteAddr, d.Payload)
			server := servers[string(d.RemoteAddr.IP)]
			if server == nil {
				var err error
				var settings struct {
					Device struct {
						Hostname string
					}
				}
				var citD struct {
					Blk []struct {
						I int
						D string
					}
					Sen []struct {
						I int
						D string
						U string
						L int
					}
				}
				server = new(coiotServer)
				err = smarthome.HTTPtoJSON(fmt.Sprintf("http://%s/settings", d.RemoteAddr.IP), &settings)
				if err != nil {
					log.Printf("Error getting settings from %s: %v", d.RemoteAddr.IP, err)
					continue
				}
				if settings.Device.Hostname == "" {
					log.Printf("No hostname found from %s", d.RemoteAddr.IP)
					continue
				}
				server.Hostname = settings.Device.Hostname

				err = smarthome.HTTPtoJSON(fmt.Sprintf("http://%s/cit/d", d.RemoteAddr.IP), &citD)
				if err != nil {
					log.Printf("Error getting CoIoT description from %s: %v", d.RemoteAddr.IP, err)
					continue
				}
				for _, blk := range citD.Blk {
					server.Blocks = append(server.Blocks, coiotBlock{ID: blk.I, Name: blk.D})
				}
				for _, sen := range citD.Sen {
					var state coiotState
					for _, block := range server.Blocks {
						if block.ID == sen.L {
							state.Block = block.Name
						}
					}
					state.ID = sen.I
					state.Name = sen.D
					state.Unit = sen.U
					server.States = append(server.States, state)
				}
				// log.Printf("Found new device %s: %+v", d.RemoteAddr.IP, server)
				servers[string(d.RemoteAddr.IP)] = server
			}
			log.Printf("Packet from %s (%s)", d.RemoteAddr.IP, server.Hostname)
			var citS struct {
				G [][]interface{}
			}
			err = json.Unmarshal(d.Payload, &citS)
			if err != nil {
				log.Printf("Error processing /cit/s from %s: %s", server.Hostname, err)
				continue
			}
			for _, val := range citS.G {
				id := int(val[1].(float64))
				value := val[2]
				for _, state := range server.States {
					if id == state.ID {
						topic := fmt.Sprintf("%s/%s/%s", server.Hostname, state.Block, state.Name)
						msg := fmt.Sprint(value)
						if state.Unit != "" {
							topic = fmt.Sprintf("%s/%s", topic, state.Unit)
							msg = fmt.Sprintf("%f", value)
							for msg[len(msg)-1] == '0' && msg[len(msg)-2] != '.' {
								msg = msg[0 : len(msg)-1]
							}

						}
						topic = fmt.Sprintf("coiot/%s", topic)
						mqtt.Publish(topic, msg)
						_ = mqtt
						break
					}
				}
			}
		}
	}
}
