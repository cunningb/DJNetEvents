package models

import (
	"time"

	"github.com/streadway/amqp"
	"gitlab.dj/libs/djnetevents/app"
)

type Subscriber struct {
	*EventPubSub
	Queue     amqp.Queue
	Listeners map[string]func(body []byte)
}

func (sub *Subscriber) reconnectSub() error {
	err := sub.Reconnect()

	if err != nil {
		return err
	}

	q, err := sub.Channel.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)

	if err != nil {
		sub.Logger.Warnf("Failed to declare queue: %s", err)
		return err
	}

	sub.Queue = q

	for routeKey, consumer := range sub.Listeners {
		if err := sub.Bind(routeKey, consumer); err != nil {
			return err
		}

		sub.Logger.Infof("Re-subscribed consumer for '%s'.'%s'", sub.ExchangeName, routeKey)
	}

	return nil
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

	sub := &Subscriber{base, q, make(map[string]func(body []byte))}

	if err != nil {
		base.Close()
		base.Logger.Warnf("Failed to declare queue: %s", err)
		return nil, err
	}

	onClose := make(chan *amqp.Error)
	sub.Channel.NotifyClose(onClose)

	// Start listening to close signal
	go func() {
		for err := range onClose {
			if err != nil {
				sub.Logger.Warnf("Subscriber Channel/Connection closed: %s", err)
				if sub.State != RECONNECTING {
					sub.State = RECONNECTING
					go func() {

						for {
							err := sub.reconnectSub()

							if err == nil {
								return
							}

							if sub.State != RECONNECTING {
								return
							}

							time.Sleep(time.Duration(app.Config.ReconnectSec) * time.Second)
						}
					}()
				}
			}
		}
	}()

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

	sub.Listeners[routeKey] = consumer

	sub.Logger.Debugf("Registered consumer for '%s'.'%s': Queue '%s'", sub.ExchangeName, routeKey, sub.Queue.Name)
	return nil
}
