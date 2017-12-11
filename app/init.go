package app

import (
	"fmt"
)

func Initialize() error {
	return loadConfigs()
}

func loadConfigs() error {
	fmt.Println("Reading RabbitMQ config")
	if err := LoadConfig("../"); err != nil {
		return fmt.Errorf("Invalid application configuration: %s", err)
	}

	return nil
}
