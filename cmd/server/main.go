package main

import (
	"fmt"
	"github.com/eampleev23/gophkeeper2.git/internal/logger"
	"github.com/eampleev23/gophkeeper2.git/internal/server_config"
	"log"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	serv_config := server_config.NewServerConfig()
	logger, err := logger.NewZapLogger(serv_config.LogLevel)
	if err != nil {
		return fmt.Errorf("failed to create zap logger: %w", err)
	}
}
