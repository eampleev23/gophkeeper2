package models

// UserRegReq - модель запроса на регистрацию.
type UserRegReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
