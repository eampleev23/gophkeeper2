package server_config

import (
	"flag"
	"os"
	"strconv"
	"time"
)

type ServerConfig struct {
	RunAddr    string // Адрес слушателя, например :8083 или 127.0.0.1:8443
	LogLevel   string
	DBDSN      string
	TokenExp   time.Duration
	SecretKey  string
	TLSEnabled bool // false = HTTP (за nginx), true = TLS локально
}

func NewServerConfig() *ServerConfig {
	servConf := &ServerConfig{
		TokenExp: time.Hour * 24 * 30, // Время сколько не истекает авторизация
	}
	servConf.SetValues()
	return servConf
}

func (c *ServerConfig) SetValues() {
	// По умолчанию TLS включён (локальная разработка на 8443)
	c.TLSEnabled = true

	flag.StringVar(&c.RunAddr, "a", "127.0.0.1:8443", "Set listening address and port for server")
	flag.StringVar(&c.LogLevel, "l", "debug", "logger level")
	flag.StringVar(&c.DBDSN, "d", "postgresql://gopher:gopher@localhost:5432/gophkeeper2?sslmode=disable", "postgres database")
	flag.StringVar(&c.SecretKey, "s", "e4853f5c4810101e88f1898db21c15d3", "server's secret key for authorization")
	flag.Parse()

	// Переменные окружения (деплой: DATABASE_URL, PORT, SECRET_KEY; TLS_ENABLED=false для HTTP за nginx)
	if envDB := os.Getenv("DATABASE_URL"); envDB != "" {
		c.DBDSN = envDB
	} else if envDB := os.Getenv("DATABASE_URI"); envDB != "" {
		c.DBDSN = envDB
	}
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 {
			c.RunAddr = ":" + strconv.Itoa(p)
			c.TLSEnabled = false
		}
	}
	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		c.RunAddr = envRunAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		c.LogLevel = envLogLevel
	}
	if envSecretKey := os.Getenv("SECRET_KEY"); envSecretKey != "" {
		c.SecretKey = envSecretKey
	}
	if envTLS := os.Getenv("TLS_ENABLED"); envTLS == "false" || envTLS == "0" {
		c.TLSEnabled = false
	} else if envTLS == "true" || envTLS == "1" {
		c.TLSEnabled = true
	}
}
