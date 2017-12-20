package main

import (
	"encoding/json"
	"fmt"

	"time"

	"gitlab.dj/libs/djnetevents/app"
	"gitlab.dj/libs/djnetevents/models"
)

func main() {
	app.Initialize(&app.RbmqConfig{
		Url:          "amqp://localhost",
		ReconnectSec: 5,
	})
	testProxyJoin()
}

type ProxyJoinEvent struct {
	Server     string `json:"server"`
	ServerType string `json:"stype"`
	Name       string `json:"pname"`
}

func testProxyJoin() {
	exchange := "djnetworkessentials.events"
	routeKey := "netplayerproxyjoinevent"

	pub, _ := models.NewPublisher(exchange)

	sub, _ := models.NewSubscriber(exchange)

	sub.Bind(routeKey, func(body []byte) {
		event := &ProxyJoinEvent{}
		json.Unmarshal(body, event)
		fmt.Println("Recieved Event ", event)
	})

	i := 0
	for {
		jsonBody, _ := json.Marshal(&ProxyJoinEvent{
			Server:     "TestServer",
			ServerType: "TestServerType",
			Name:       fmt.Sprintf("Test#%d", i),
		})
		i++

		pub.Publish(jsonBody, routeKey)
		time.Sleep(5 * time.Second)
	}

}
