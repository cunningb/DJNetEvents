package models

import (
	"encoding/json"
	"time"

	"github.com/streadway/amqp"
	"gitlab.dj/libs/djnetevents/app"
)

type Publisher struct {
	*EventPubSub
}

func NewPublisher(exchangeName string) (*Publisher, error) {
	base, err := NewPubSub(exchangeName)
	if err != nil {
		return nil, err
	}

	pub := &Publisher{base}
	pub.registerNotifyClose()

	return pub, nil
}

func (pub *Publisher) registerNotifyClose() {
	onClose := make(chan *amqp.Error)
	pub.Channel.NotifyClose(onClose)

	// Start listening to close signal
	go func() {
		for err := range onClose {
			if err != nil {
				pub.Logger.Warnf("Publisher Channel/Connection closed: %s", err)
				pub.startReconnectionTask()
			}
		}
	}()

	pub.Logger.Info("Publisher notifyClose listener registered")
}

func (pub *Publisher) startReconnectionTask() {
	if pub.State != RECONNECTING {
		pub.State = RECONNECTING
		pub.Logger.Info("Starting publisher reconnection task")
		go func() {

			for {
				err := pub.Reconnect()

				if err == nil {
					pub.registerNotifyClose()
					return
				}

				if pub.State != RECONNECTING {
					return
				}

				time.Sleep(time.Duration(app.Config.ReconnectSec) * time.Second)
			}

		}()
	}
}

// Publish a slice of bytes to a specific routeKey
func (ps *Publisher) Publish(data []byte, routeKey string) error {
	err := ps.Channel.Publish(
		ps.ExchangeName, // exchange
		routeKey,        // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(data),
		})

	if err == nil {
		ps.Logger.Debugf("Publishing %d bytes to routekey %s", len(data), routeKey)
	} else {
		ps.Logger.Warnf("Failed to publish %d bytes to routekey %s (State: %s): %s", len(data), routeKey, ps.State, err)
	}

	return err
}

// Publish an object to a routeKey. This will first marshal the object to JSON, and then publish it using the `(Publisher) Publish([]byte,string) (error)` function
func (ps *Publisher) PublishJson(data interface{}, routeKey string) error {
	json, err := json.Marshal(data)

	if err != nil {
		return err
	}

	return ps.Publish(json, routeKey)
}
