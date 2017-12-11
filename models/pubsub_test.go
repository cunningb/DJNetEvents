package models

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"gitlab.dj/libs/djnetevents/app"
	"testing"
)

type TestEvent struct {
	Server     string `json:"server"`
	ServerType string `json:"stype"`
	Name       string `json:"pname"`
}

func TestPubSub(t *testing.T) {
	fmt.Println("Starting PubSub test")
	app.Initialize()

	exchange := "gotestExchange"
	routeKey := "gotest"

	pub, err := NewPublisher(exchange)
	assert.Nil(t, err)
	assert.NotNil(t, pub)

	sub, err := NewSubscriber(exchange)
	assert.Nil(t, err)
	assert.NotNil(t, sub)

	msg := []byte("FirstEventTest")
	forever := make(chan bool)

	fmt.Println("Binding consumer for raw bytes")
	err = sub.Bind(routeKey, func(body []byte) {
		fmt.Println("Recieved ", string(body))
		assert.Equal(t, msg, body)
		forever <- true
	})
	assert.Nil(t, err)

	fmt.Println("Publishing ", string(msg), " on ", routeKey)
	err = pub.Publish(msg, routeKey)
	assert.Nil(t, err)

	<-forever

	pub.Close()
	sub.Close()

	fmt.Println("Testing offline pub/sub")
	err = pub.Publish(msg, routeKey)
	assert.NotNil(t, err)

	err = sub.Bind(routeKey, func(body []byte) {
		assert.Equal(t, msg, body)
	})

	assert.NotNil(t, err)
}

func TestEventPubSub(t *testing.T) {
	fmt.Println("Starting Event test")
	app.Initialize()

	exchange := "goEventExchange"
	routeKey := "gotest"

	pub, err := NewPublisher(exchange)
	assert.Nil(t, err)
	assert.NotNil(t, pub)

	sub, err := NewSubscriber(exchange)
	assert.Nil(t, err)
	assert.NotNil(t, sub)

	msg := &TestEvent{
		Server:     "testServer",
		ServerType: "testServerType",
		Name:       "TestPlayerName",
	}

	forever := make(chan bool)

	err = sub.Bind(routeKey, func(body []byte) {
		event := &TestEvent{}
		err := json.Unmarshal(body, event)
		assert.Nil(t, err)

		fmt.Println("Recieved Event ", event)
		assert.Equal(t, msg, event)
		forever <- true
	})
	assert.Nil(t, err)

	jsonBody, err := json.Marshal(msg)
	assert.Nil(t, err)

	err = pub.Publish(jsonBody, routeKey)
	assert.Nil(t, err)

	<-forever
}
