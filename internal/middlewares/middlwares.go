package middlewares

// resultMessage - структура для возврата json ответа более детализированного, чем просто статус.
type resultMsg struct {
	IsError       bool
	ResultMessage string `json:"result_message"`
}
