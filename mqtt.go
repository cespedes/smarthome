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

	if m.root != "" {
		topic = fmt.Sprintf("%s/%s", m.root, topic)
	}

	log.Printf("MQTT: Publishing %s=%s\n", topic, payload)
	err = m.client.Publish(context.Background(), &mqtt.Message{
		Topic:   topic,
		Payload: []byte(payload),
	})
	if err != nil {
		panic(err)
	}
}

func (m *MQTTClient) Subscribe(topic string) chan *mqtt.Message {
	ch := make(chan *mqtt.Message)
	m.mux.HandleFunc(topic, func(m *mqtt.Message) {
		ch <- m
	})

	_, err := m.client.Subscribe(context.Background(), mqtt.Subscription{Topic: topic})
	if err != nil {
		panic(err)
	}

	return ch
}
