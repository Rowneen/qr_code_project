package handlers

import (
	"encoding/json"
	"net/http"
	"qr_code/internal/config"
	"regexp"
)

// запрос
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// ответ
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// чек через регулярку полей ввода (защита от хуйни и sql инъекций)
func isValidInput(input string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9@.\-_]{3,50}$`, input)
	return matched
}

// сами хандлеры обработчики
func handler_auth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		response := AuthResponse{
			Success: false,
			Message: "Only POST method allowed",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	var authRequest AuthRequest
	decoder := json.NewDecoder(r.Body)
	// проверка на невалидно переданный json (не получается распарсить)
	if err := decoder.Decode(&authRequest); err != nil {
		response := AuthResponse{
			Success: false,
			Message: "Invalid JSON format",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// проверка полей на пустоту
	if authRequest.Login == "" || authRequest.Password == "" {
		response := AuthResponse{
			Success: false,
			Message: "Login and password are required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// проверка полей на валидность по символам
	if !isValidInput(authRequest.Login) || !isValidInput(authRequest.Password) {
		response := AuthResponse{
			Success: false,
			Message: "Login or password contains invalid characters. Only letters, numbers, @, ., -, _ are allowed (3-50 characters)",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// заглушка
	if authRequest.Login == "admin" && authRequest.Password == "password" {
		response := AuthResponse{
			Success: true,
			Message: "Authentication successful",
		}
		json.NewEncoder(w).Encode(response)
	} else {
		response := AuthResponse{
			Success: false,
			Message: "Invalid login or password",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
	}
}

// регистрация хандлеров и запуск
func RegisterHTTPHandlers() {
	cfg := config.Get()
	// конфиг сервера
	server := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      nil,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}
	// запуск обработчиков на функции
	http.HandleFunc("/auth", handler_auth)

	server.ListenAndServe()
}
