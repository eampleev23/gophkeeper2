package store

import (
	"fmt"
	"github.com/eampleev23/gophkeeper2.git/internal/logger"
	"github.com/eampleev23/gophkeeper2.git/internal/server_config"
)

type Store interface {
	// DBConnClose закрывает соединение с базой данных
	DBConnClose() (err error)
}

func NewStorage(serv_conf *server_config.ServerConfig, logger *logger.ZapLog) (Store, error) {
	s, err := NewDBStore(serv_conf, logger)
	if err != nil {
		return nil, fmt.Errorf("error creating new db store: %w", err)
	}
	logger.ZL.Debug("DB store created success..")
	return s, nil
}
