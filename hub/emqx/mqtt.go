package emqx

import (
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MQTTClient struct {
	Client mqtt.Client
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
	return &MQTTClient{Client: client}
}

func (m *MQTTClient) Subscribe(topic string, qos byte, handler mqtt.MessageHandler) {
	token := m.Client.Subscribe(topic, qos, handler)
	token.Wait()
	if token.Error() != nil {
		log.Fatal(token.Error())
	}
	fmt.Println("Subscribed to:", topic)
}

func (m *MQTTClient) Publish(topic string, qos byte, retained bool, payload interface{}) {
	token := m.Client.Publish(topic, qos, retained, payload)
	token.Wait()
	fmt.Println("Published:", payload)
}


func (m *MQTTClient) Disconnect() {
	m.Client.Disconnect(250)
	fmt.Println("Disconnected")
}

// Example publisher loop (optional utility)
func (m *MQTTClient) PublishLoop(topic string, count int, delay time.Duration) {
	for i := 1; i <= count; i++ {
		msg := fmt.Sprintf("Score update %d", i)
		m.Publish(topic, 0, false, msg)
		time.Sleep(delay)
	}
}
