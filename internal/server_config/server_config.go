package server_config

import (
	"flag"
	"os"
	"time"
)

type ServerConfig struct {
	RunAddr   string
	LogLevel  string
	DBDSN     string
	TokenExp  time.Duration
	SecretKey string
}

func NewServerConfig() *ServerConfig {
	servConf := &ServerConfig{
		TokenExp: time.Hour * 24 * 30, // Время сколько не истекает авторизация
	}
	servConf.SetValues()
	return servConf
}

func (c *ServerConfig) SetValues() {
	// регистрируем переменную flagRunAddr как аргумент -a со значением по умолчанию localhost:8080
	//flag.StringVar(&c.RunAddr, "a", "localhost:8080", "Set listening address and port for server")
	flag.StringVar(&c.RunAddr, "a", "0.0.0.0:8080", "Set listening address and port for server")
	// регистрируем уровень логирования
	flag.StringVar(&c.LogLevel, "l", "debug", "logger level")
	// принимаем строку подключения к базе данных
	flag.StringVar(&c.DBDSN, "d", "", "postgres database")
	// принимаем секретный ключ сервера для авторизации
	flag.StringVar(&c.SecretKey, "s", "e4853f5c4810101e88f1898db21c15d3", "server's secret key for authorization")

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		c.RunAddr = envRunAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		c.LogLevel = envLogLevel
	}
	if envDBDSN := os.Getenv("DATABASE_URI"); envDBDSN != "" {
		c.DBDSN = envDBDSN
	}
	if envSecretKey := os.Getenv("SECRET_KEY"); envSecretKey != "" {
		c.SecretKey = envSecretKey
	}
}
