package models

import (
	"github.com/streadway/amqp"
	"gitlab.dj/libs/djnetevents/app"
)

type EventPubSub struct {
	Conn         *amqp.Connection
	Channel      *amqp.Channel
	ExchangeName string
	Logger       app.Logger
}

func (ps *EventPubSub) Close() {
	if ps.Channel != nil {
		ps.Channel.Close()
	}

	if ps.Conn != nil {
		ps.Conn.Close()
	}
}

func NewPubSub(exchangeName string) (*EventPubSub, error) {
	pub := &EventPubSub{}
	pub.ExchangeName = exchangeName
	pub.Logger = app.DefaultLogger

	conn, err := amqp.Dial(app.Config.Url)
	if err != nil {
		conn.Close()
		pub.Logger.Warnf("Failed to establish rbmq connection with %s: %s\n", app.Config.Url, err)
		return nil, err
	}

	pub.Conn = conn

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		ch.Close()
		pub.Logger.Warnf("Failed to create channel: %s\n", err)
		return nil, err
	}

	pub.Channel = ch

	err = pub.Channel.ExchangeDeclare(
		pub.ExchangeName, // name
		"direct",         // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)

	if err != nil {
		pub.Close()
		pub.Logger.Warnf("Failed to bind exchange %s: %s\n", pub.ExchangeName, err)
		return nil, err
	}

	return pub, nil
}
