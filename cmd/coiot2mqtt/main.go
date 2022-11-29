package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

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
	conf := readConfig()

	if conf.Debug {
		log.Println("Creating MQTT client...")
	}
	mqtt, err := smarthome.NewMQTTClient(conf.MQTTServer, conf.MQTTPrefix)
	if err != nil {
		log.Fatal(err)
	}

	if conf.Debug {
		log.Println("Listening to CoIoT packets...")
	}
	coiot, err := smarthome.CoIoTinit()
	if err != nil {
		panic(err)
	}

	var lock = sync.RWMutex{}
	servers := make(map[string]*coiotServer)

	mqttChan := mqtt.Subscribe("+/cmd")
	go func() {
		for {
			msg := <-mqttChan
			parts := strings.Split(msg.Topic, "/")
			if len(parts) < 2 {
				log.Printf("Internal error: received message with topic %q", msg.Topic)
				continue
			}
			shellyID := parts[len(parts)-2]
			var ipAddr string
			lock.RLock()
			for key, server := range servers {
				if server.Hostname == shellyID {
					ipAddr = key
					break
				}
			}
			lock.RUnlock()
			if ipAddr == "" {
				log.Printf("Received command for unknown Shelly device %q", shellyID)
				continue
			}
			payload := string(msg.Payload)
			parts = strings.Split(payload, " ")
			if len(parts) < 1 {
				log.Printf("Received incorrect command for Shelly device %q: %q", shellyID, payload)
				continue
			}
			cmd := parts[0]
			args := parts[1:]
			cmdSyntaxError := func() {
				log.Printf("CMD: syntax error (sent %q to %q)", payload, shellyID)
			}
			switch {
			case cmd == "reboot":
				if len(args) != 0 {
					cmdSyntaxError()
					continue
				}
				log.Printf("CMD: would reboot %s (%s) now (unimplemented)", shellyID, ipAddr)
			case strings.HasPrefix(cmd, "relay/") && len(cmd) > 6:
				relayID, err := strconv.Atoi(cmd[6:])
				if err != nil || len(args) < 1 || len(args) > 2 ||
					(args[0] != "on" && args[0] != "off" && args[0] != "toggle") ||
					(len(args) == 2 && !strings.HasPrefix(args[1], "timer=")) {
					cmdSyntaxError()
					continue
				}
				timer := 0
				if len(args) == 2 {
					timer, err = strconv.Atoi(args[1][6:])
					if err != nil {
						cmdSyntaxError()
						continue
					}
				}
				log.Printf("CMD: %s (%s): changing relay %d to %s (timer=%d)", shellyID, ipAddr, relayID, args[0], timer)
				http.Get(fmt.Sprintf("http://%s/relay/%d?turn=%s&timer=%d", ipAddr, relayID, args[0], timer))

				// roller/<id> {open | close | stop | <pos> | calibrate}

			case strings.HasPrefix(cmd, "roller/") && len(cmd) > 7:
				rollerID, err := strconv.Atoi(cmd[7:])
				if err != nil || len(args) != 1 {
					cmdSyntaxError()
					continue
				}
				pos := -1
				pos, err = strconv.Atoi(args[0])
				if err != nil && args[0] != "open" && args[0] != "close" &&
					args[0] != "stop" && args[0] != "calibrate" {
					cmdSyntaxError()
					continue
				}
				switch {
				case err == nil:
					log.Printf("CMD: %s (%s): positioning roller %d to %d%%", shellyID, ipAddr, rollerID, pos)
					http.Get(fmt.Sprintf("http://%s/roller/%d?go=to_pos&roller_pos=%d", ipAddr, rollerID, pos))
				case args[0] == "calibrate":
					log.Printf("CMD: %s (%s): calibrating roller %d", shellyID, ipAddr, rollerID)
					http.Get(fmt.Sprintf("http://%s/roller/%d/calibrate", ipAddr, rollerID))
				default:
					log.Printf("CMD: %s (%s): changing roller %d to %s", shellyID, ipAddr, rollerID, args[0])
					http.Get(fmt.Sprintf("http://%s/roller/%d?go=%s", ipAddr, rollerID, args[0]))
				}
			default:
				log.Printf("CMD: shellyID=%q IP=%q payload=%q", shellyID, ipAddr, payload)
			}
		}
	}()

	for {
		d := coiot.Read()
		if conf.Debug {
			log.Printf("CoIoT packet from %s: %q\n", d.RemoteAddr, d.Payload)
		}
		if d.Code[0] != 0 || d.Code[1] != 30 {
			// this is not a CoIoT packet
			continue
		}
		// log.Printf("CoIoT packet from %s: %q\n", d.RemoteAddr, d.Payload)
		lock.RLock()
		server := servers[d.RemoteAddr.IP.String()]
		lock.RUnlock()
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
			lock.Lock()
			servers[d.RemoteAddr.IP.String()] = server
			lock.Unlock()
		}
		if conf.Debug {
			log.Printf("Valid CoIoT packet from %s (%s)", d.RemoteAddr.IP, server.Hostname)
		}
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
					mqtt.Publish(topic, msg)
					_ = mqtt
					break
				}
			}
		}
	}
}
