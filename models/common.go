package models

import (
	"github.com/streadway/amqp"
	"gitlab.dj/libs/djnetevents/app"
)

type PubSubState uint32

const (
	INIT PubSubState = iota
	CONNECTED
	RECONNECTING
	SHUTDOWN
)

type EventPubSub struct {
	Conn         *amqp.Connection
	Channel      *amqp.Channel
	ExchangeName string
	Logger       app.Logger
	State        PubSubState
}

func (ps *EventPubSub) Close() {
	ps.State = SHUTDOWN

	if ps.Channel != nil {
		ps.Channel.Close()
	}

	if ps.Conn != nil {
		ps.Conn.Close()
	}
}

func (ps *EventPubSub) Reconnect() error {
	if ps.State == SHUTDOWN {
		ps.Logger.Infof("Connection attempt aborted")
		return nil
	}

	conn, err := amqp.Dial(app.Config.Url)
	if err != nil {
		ps.Logger.Warnf("Failed to establish rbmq connection with %s: %s", app.Config.Url, err)
		return err
	}

	ps.Conn = conn

	ch, err := conn.Channel()
	if err != nil {
		ps.Logger.Warnf("Failed to create channel: %s", err)
		return err
	}

	ps.Channel = ch

	err = ps.Channel.ExchangeDeclare(
		ps.ExchangeName, // name
		"topic",         // type
		false,           // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)

	if err != nil {
		ps.Logger.Warnf("Failed to bind exchange %s: %s", ps.ExchangeName, err)
		return err
	}

	ps.State = CONNECTED
	ps.Logger.Infof("Successfully connected to '%s', Exchange '%s'", app.Config.Url, ps.ExchangeName)

	return nil
}

func NewPubSub(exchangeName string) (*EventPubSub, error) {
	pub := &EventPubSub{}
	pub.ExchangeName = exchangeName
	pub.Logger = app.DefaultLogger

	err := pub.Reconnect()

	if err != nil {
		pub.Close()
		return nil, err
	}

	return pub, nil
}
