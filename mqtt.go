package smarthome

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"time"

	"github.com/at-wat/mqtt-go"
)

type MQTTClient struct {
	client mqtt.ReconnectClient
	mux    *mqtt.ServeMux
	root   string
}

func NewMQTTClient(addr string) (*MQTTClient, error) {
	var m MQTTClient
	var err error

	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	root := ""
	if len(u.Path) > 0 && u.Path[0] == '/' {
		root = u.Path[1:]
	}

	// log.Printf("MQTT: Connecting to server %q...", addr)
	m.client, err = mqtt.NewReconnectClient(&mqtt.URLDialer{URL: addr})
	if err != nil {
		return nil, err
	}

	rand.Seed(time.Now().UnixNano())
	_, err = m.client.Connect(context.Background(), fmt.Sprint(rand.Uint64()))
	if err != nil {
		return nil, err
	}
	// log.Println("MQTT: Connected.")

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

	// log.Printf("MQTT: Publishing %s=%s\n", topic, payload)
	err = m.client.Publish(context.Background(), &mqtt.Message{
		Topic:   topic,
		Payload: []byte(payload),
	})
	if err != nil {
		panic(err)
	}
}

func (m *MQTTClient) Subscribe(topic string) chan *mqtt.Message {
	if m.root != "" {
		topic = fmt.Sprintf("%s/%s", m.root, topic)
	}

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
