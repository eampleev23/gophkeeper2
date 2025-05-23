package handlers

import (
	"go.uber.org/zap"
	"net/http"
	"strings"
)

func (handlers *Handlers) Registration(responseWriter http.ResponseWriter, request *http.Request) {

	// Проверяем, что пришел запрос в JSON.
	contentType := request.Header.Get("Content-Type")
	supportJSON := strings.Contains(contentType, "application/json")
	if !supportJSON {
		responseWriter.WriteHeader(http.StatusBadRequest)
		return
	}

	_, isAuth, err := handlers.GetUserID(request)
	if err != nil {
		handlers.l.ZL.Error("GetUserID failed", zap.Error(err))
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	if isAuth {
		handlers.l.ZL.Error("already authorized user is going to register")
		responseWriter.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Получаем данные в случае корректного запроса
}
