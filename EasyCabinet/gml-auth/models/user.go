package models

// User — запись пользователя в JSON базе данных
type User struct {
	UUID        string `json:"uuid"`
	Login       string `json:"login"`
	Password    string `json:"password"`
	IsSlim      bool   `json:"is_slim"`
	Blocked     bool   `json:"blocked"`
	BlockReason string `json:"block_reason"`
}

// Database — структура JSON файла
type Database struct {
	Users []User `json:"users"`
}

// AuthRequest — запрос от GML Launcher
type AuthRequest struct {
	Login    string `json:"Login"`
	Password string `json:"Password"`
	Totp     string `json:"Totp"`
}

// AuthResponse — ответ при успешной авторизации (200)
type AuthResponse struct {
	Login    string `json:"Login"`
	UserUuid string `json:"UserUuid"`
	IsSlim   bool   `json:"IsSlim"`
	Message  string `json:"Message"`
}

// ErrorResponse — ответ при ошибке (401, 403, 404)
type ErrorResponse struct {
	Message string `json:"Message"`
}

// CreateUserRequest — запрос для создания пользователя через admin API
type CreateUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
	IsSlim   bool   `json:"is_slim"`
}

// BlockRequest — запрос для блокировки пользователя
type BlockRequest struct {
	Reason string `json:"reason"`
}

// WebErrorResponse — формат ошибки для GML Launcher web-панели (ожидает errors[])
type WebErrorResponse struct {
	Errors []string `json:"errors"`
}
