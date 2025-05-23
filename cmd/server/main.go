package main

import (
	"fmt"
	"log"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	c, err := server_config.NewConfig()
	if err != nil {
		return fmt.Errorf("failed to initialize a new config: %w", err)
	}
}
