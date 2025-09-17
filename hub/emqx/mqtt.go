package emqx

import (
	"fmt"
	"log"
	"time"

     "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	client mqtt.Client
}

func NewClient(broker, clientID string, defaultHandler mqtt.MessageHandler) *MQTTClient {
	opts := mqtt.NewClientOptions().AddBroker(broker).SetClientID(clientID)

	if defaultHandler != nil {
		opts.SetDefaultPublishHandler(defaultHandler)
	}

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	fmt.Printf("Connected [%s] to broker %s\n", clientID, broker)
	return &MQTTClient{client: client}
}

func (m *MQTTClient) Subscribe(topic string, qos byte, handler mqtt.MessageHandler) error {
	token := m.client.Subscribe(topic, qos, handler)
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("subscribe timeout for topic %q", topic)
	}
	if err := token.Error(); err != nil {
		return err
	}
	return nil
}

func (m *MQTTClient) Publish(topic string, qos byte, retained bool, payload interface{}) error {
	token := m.client.Publish(topic, qos, retained, payload)
	if !token.WaitTimeout(5 * time.Second) {
		return fmt.Errorf("publish timeout for topic %q", topic)
	}
	if err := token.Error(); err != nil {
		return err
	}
	return nil
}


func (m *MQTTClient) Disconnect() {
	m.client.Disconnect(250)
	fmt.Println("Disconnected")
}
