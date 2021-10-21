package smarthome

import (
	"context"
	"fmt"
	"log"

	"github.com/at-wat/mqtt-go"
)

type MQTTClient struct {
	client mqtt.ReconnectClient
	mux    *mqtt.ServeMux
	root   string
}

func NewMQTTClient(addr string, root string) (*MQTTClient, error) {
	var m MQTTClient
	var err error

	log.Printf("MQTT: Connecting to server %q...", addr)
	m.client, err = mqtt.NewReconnectClient(&mqtt.URLDialer{URL: addr})
	if err != nil {
		return nil, err
	}

	_, err = m.client.Connect(context.Background(), "")
	if err != nil {
		return nil, err
	}
	log.Println("MQTT: Connected.")

	m.mux = &mqtt.ServeMux{}
	m.client.Handle(m.mux)

	m.root = root

	return &m, nil
}

func (m *MQTTClient) Publish(topic string, payload string) {
	var err error

	log.Printf("MQTT: Publishing %s/%s=%s\n", m.root, topic, payload)
	err = m.client.Publish(context.Background(), &mqtt.Message{
		Topic:   fmt.Sprintf("%s/%s", m.root, topic),
		Payload: []byte(payload),
	})
	if err != nil {
		panic(err)
	}
}
