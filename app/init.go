package app

// RbmqConfig for RabbitMq configurations such as connection url
type RbmqConfig struct {
	// Url for rabbitmq connection, must contain 'amqp://' or 'amqps://'
	Url string
	// How frequently to try and reconnect to RabbitMq when connection is lost
	ReconnectSec int64
	// Logger to use, optional
	Logger Logger
}

var Config *RbmqConfig

func Initialize(config *RbmqConfig) error {
	Config = config

	if Config.Logger == nil {
		Config.Logger = DefaultLogger
	}

	return nil
}
