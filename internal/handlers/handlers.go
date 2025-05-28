package handlers

import (
	"encoding/json"
	"github.com/eampleev23/gophkeeper2.git/internal/auth"
	"github.com/eampleev23/gophkeeper2.git/internal/logger"
	"github.com/eampleev23/gophkeeper2.git/internal/server_config"
	"github.com/eampleev23/gophkeeper2.git/internal/store"
	"net/http"
)

type Handlers struct {
	store    store.Store
	servConf *server_config.ServerConfig
	logger   *logger.ZapLog
	auth     *auth.Authorizer
}

func NewHandlers(
	store store.Store,
	servConf *server_config.ServerConfig,
	logger *logger.ZapLog,
	auth *auth.Authorizer,
) (
	*Handlers,
	error) {
	return &Handlers{
		store:    store,
		servConf: servConf,
		logger:   logger,
		auth:     auth,
	}, nil
}

// resultMessage - структура для возврата json ответа более детализированного, чем просто статус.
type resultMsg struct {
	IsError       bool
	ResultMessage string `json:"result_message"`
}

func sendResponse(
	isError bool,
	mg string,
	status int,
	responseWriter http.ResponseWriter,
) (err error) {
	resultMsg := resultMsg{IsError: isError, ResultMessage: mg}
	msg, _ := json.Marshal(resultMsg)
	responseWriter.WriteHeader(status)
	responseWriter.Write(msg)
	return nil
}
