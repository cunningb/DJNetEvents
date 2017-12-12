package models

import (
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

	onClose := make(chan *amqp.Error)
	pub.Channel.NotifyClose(onClose)

	// Start listening to close signal
	go func() {
		for err := range onClose {
			if err != nil {
				pub.Logger.Warnf("Publisher Channel/Connection closed: %s", err)
				if pub.State != RECONNECTING {
					pub.State = RECONNECTING
					go func() {

						for {
							err := pub.Reconnect()

							if err == nil {
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
		}
	}()

	return pub, nil
}

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
		ps.Logger.Warnf("Failed to publish %d bytes to routekey %s: %s", len(data), routeKey, err)
	}

	return err
}
