package server_config

import (
	"flag"
	"os"
	"time"
)

type ServerConfig struct {
	RunAddr  string
	LogLevel string
	DBDSN    string
	TokenExp time.Duration
}

func NewServerConfig() *ServerConfig {
	servConf := &ServerConfig{
		TokenExp: time.Hour * 24 * 30, // Время сколько не истекает авторизация
	}
	return servConf
}

func (c *ServerConfig) SetValues() {
	// регистрируем переменную flagRunAddr как аргумент -a со значением по умолчанию localhost:8080
	//flag.StringVar(&c.RunAddr, "a", "localhost:8080", "Set listening address and port for server")
	flag.StringVar(&c.RunAddr, "a", "0.0.0.0:8080", "Set listening address and port for server")
	// регистрируем уровень логирования
	flag.StringVar(&c.LogLevel, "l", "debug", "logger level")
	// принимаем строку подключения к базе данных
	flag.StringVar(&c.DBDSN, "d", "postgresql://postgres:j0Wam3ibcT4KnGWUWuabEpuUmzL@212.193.48.196:5432/template1", "postgres database")

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		c.RunAddr = envRunAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		c.LogLevel = envLogLevel
	}
	if envDBDSN := os.Getenv("DATABASE_URI"); envDBDSN != "" {
		c.DBDSN = envDBDSN
	}
}
