package models

import (
	"github.com/streadway/amqp"
)

type Publisher struct {
	*EventPubSub
}

func NewPublisher(exchangeName string) (*Publisher, error) {
	base, err := NewPubSub(exchangeName)
	if err != nil {
		return nil, err
	}

	return &Publisher{base}, nil
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
