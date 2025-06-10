package handlers

import (
	"encoding/json"
	"github.com/eampleev23/gophkeeper2.git/internal/models"
	"net/http"
)

/*
На вход хэндлер ожидает json такого формата:
{
    "login": "<login>",
    "password": "<password>"
}
*/

func (handlers *Handlers) Login(responseWriter http.ResponseWriter, gotRequest *http.Request) {

	// Создаем модель для парсинга запроса.
	var userLoginReq models.UserLoginReq

	// Пробуем спарсить запрос в модель.
	decoder := json.NewDecoder(gotRequest.Body)
	if err := decoder.Decode(&userLoginReq); err != nil {
		sendResponse(
			true,
			"Not a valid user login request",
			http.StatusBadRequest,
			responseWriter)
		return
	}

	// Дополнительная проверка на пустые значения
	if userLoginReq.Login == "" || userLoginReq.Password == "" {
		sendResponse(
			true,
			"Login and password are required",
			http.StatusBadRequest,
			responseWriter)
		return
	}

}
