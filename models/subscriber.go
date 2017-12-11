package models

import (
	"github.com/streadway/amqp"
)

type Subscriber struct {
	*EventPubSub
	Queue amqp.Queue
}

// NewSubscriber craetes new subscriber for given exchage
func NewSubscriber(exchangeName string) (*Subscriber, error) {
	base, err := NewPubSub(exchangeName)
	if err != nil {
		return nil, err
	}

	q, err := base.Channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		base.Close()
		base.Logger.Warnf("Failed to declare queue: %s", err)
		return nil, err
	}

	sub := &Subscriber{base, q}

	return sub, nil
}

// Bind subscriber for given route key, redirecting messages to provided consumer
func (sub *Subscriber) Bind(routeKey string, consumer func(body []byte)) error {
	err := sub.Channel.QueueBind(
		sub.Queue.Name,   // queue name
		routeKey,         // route key
		sub.ExchangeName, // Exchange name
		false,
		nil,
	)

	if err != nil {
		sub.Logger.Warnf("Failed to bind queue '%s' -> '%s.%s': %s", sub.Queue.Name, sub.ExchangeName, routeKey, err)
		return err
	}

	msgs, err := sub.Channel.Consume(
		sub.Queue.Name, // Queue name
		"",             // Consumer
		true,           // Auto ack
		false,          // Exclusive
		false,          // No local
		false,          // No wait
		nil,            // Args
	)

	if err != nil {
		sub.Logger.Warnf("Failed to create consumer for queue '%s': %s", sub.Queue.Name, err)
		return err
	}

	go func() {
		for d := range msgs {
			consumer(d.Body)
		}
	}()

	sub.Logger.Debugf("Registered consumer for '%s'.'%s': Queue '%s'", sub.ExchangeName, routeKey, sub.Queue.Name)
	return nil
}
