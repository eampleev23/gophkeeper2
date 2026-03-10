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
	auth *auth.Authorizer, // Убрать *
) (*Handlers, error) {
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
	statusCode int,
	responseWriter http.ResponseWriter,
) (err error) {
	resultMsg := resultMsg{IsError: isError, ResultMessage: mg}
	msg, _ := json.Marshal(resultMsg)
	responseWriter.WriteHeader(statusCode)
	responseWriter.Write(msg)
	return nil
}

// Health — эндпоинт для проверки живости (деплой, nginx).
func (h *Handlers) Health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Root — ответ на корневой путь (как у pointscounter).
func (h *Handlers) Root(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"environment": "production",
		"message":     "Gophkeeper2 API",
		"version":     "1.0.0",
	})
}
