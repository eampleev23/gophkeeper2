package handlers

import "net/http"

/*
На вход хэндлер ожидает json такого формата:
{
    "login": "<login>",
    "password": "<password>"
}
*/

func (handlers *Handlers) Login(w http.ResponseWriter, r *http.Request) {

}
