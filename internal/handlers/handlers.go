package handlers

import (
	"github.com/eampleev23/gophkeeper2.git/internal/logger"
	"github.com/eampleev23/gophkeeper2.git/internal/server_config"
	"github.com/eampleev23/gophkeeper2.git/internal/store"
)

type Handlers struct {
	s store.Store
	c *server_config.ServerConfig
	l *logger.ZapLog
}

func NewHandlers(
	s store.Store,
	c *server_config.ServerConfig,
	l *logger.ZapLog,
) (
	*Handlers,
	error) {
	return &Handlers{
		s: s,
		c: c,
		l: l,
	}, nil
}
