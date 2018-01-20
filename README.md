### How to use
* Initialize by calling `func Initialize(config *RbmqConfig) error` inside `github.com/cunningb/djnetevents/app` package. 
* Structure of the config is as follows, logger is optional

```golang

// RbmqConfig structure for reference
// RbmqConfig for RabbitMq configurations such as connection url
type RbmqConfig struct {
	// Url for rabbitmq connection, must contain 'amqp://' or 'amqps://'
	Url string
	// How frequently to try and reconnect to RabbitMq when connection is lost
	ReconnectSec int64
	// Logger to use, optional
	Logger Logger
}

// Example initialization
app.Initialize(&app.RbmqConfig{
		Url:          "amqp://localhost",
		ReconnectSec: 5,
	})
```

* Once initialized define a structure for the event. DJNet expects the event to contain server name (json field: *server*) and server type (json field: *stype*) fields in the json. Example:

```golang
type ProxyJoinEvent struct {
	Server     string `json:"server"`
	ServerType string `json:"stype"`
	Name       string `json:"pname"`
}
```

* You need to decide on **exchange** and **route key** to send the event to. On plugin side of things, most event exchanges have the format of **PLUGIN-NAME.events**, for example DJNet events fired on the exchange `djnetworkessentials.events` while the routekey is the class of the event object, for example `netplayerproxyjoinevent` for event class `NetPlayerProxyJoinEvent`

* To Publish events, you need to initialize a *Publisher* from `github.com/cunningb/djnetevents/models` package, serialize the event into json, and send using `func (ps *Publisher) Publish(data []byte, routeKey string) error` method

```golang
import (
	"encoding/json"
	"github.com/cunningb/djnetevents/app"
	"github.com/cunningb/djnetevents/models"
)
	
	exchange := "djnetworkessentials.events"
	routeKey := "netplayerproxyjoinevent"

	pub, err := models.NewPublisher(exchange)
	// Error check, if error is thrown at this point, it will not try to reconnect
	
	jsonBody, err := json.Marshal(&ProxyJoinEvent{
			Server:     "TestServer",
			ServerType: "TestServerType",
			Name:       fmt.Sprintf("Test#%d", 123),
		})
	// Error check

	pub.Publish(jsonBody, routeKey)
```

* To listen to the event, initialize a *Subscriber* from  `github.com/cunningb/djnetevents/models`, add bind a listener using `func (sub *Subscriber) Bind(routeKey string, consumer func(body []byte)) error` function (currently supports one listener per route key for given subscriber)

```golang
import (
	"encoding/json"
	"github.com/cunningb/djnetevents/app"
	"github.com/cunningb/djnetevents/models"
)
	exchange := "djnetworkessentials.events"
	routeKey := "netplayerproxyjoinevent"

	sub, err := models.NewSubscriber(exchange)
	// Error check, if error is thrown at this point, it will not try to reconnect
	
	err := sub.Bind(routeKey, func(body []byte) {
		event := &ProxyJoinEvent{}
		json.Unmarshal(body, event)
		fmt.Println("Recieved Event ", event)
	})
	// Error check
```

* main.go have this example code working, sending events in an infinite loop