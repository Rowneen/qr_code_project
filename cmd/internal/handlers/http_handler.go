package handlers

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"qr_code/internal/config"
	"qr_code/internal/database"
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

// удалялка спецсимволов (на всякий случай)
func sanitizeInput(input string) string {
	// Удаляем все, кроме: букв, цифр, @, ., -, _
	reg := regexp.MustCompile(`[^a-zA-Z0-9@.\-_]`)
	return reg.ReplaceAllString(input, "")
}

// md5 от строки
func md5Hash(input string) string {
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
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

	// логика авторизации
	db := database.Get()
	cleanLogin, cleanPassword := sanitizeInput(authRequest.Login), sanitizeInput(authRequest.Password)

	var (
		id       int
		Login    string
		PassHash string
		FullName string
		Role     string
		GroupId  sql.NullInt64
	)

	err := db.QueryRow(
		"SELECT * FROM user WHERE Login = ? AND PassHash = ?",
		cleanLogin, md5Hash(cleanPassword),
	).Scan(&id, &Login, &PassHash, &FullName, &Role, &GroupId)

	if err != nil {
		// log.Printf("db error: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "Invalid login or password",
		})
		return
	}

	log.Printf("correct auth: %d, %s, %s, %s, %s, %d", id, Login, PassHash, FullName, Role, GroupId.Int64)
	json.NewEncoder(w).Encode(AuthResponse{
		Success: true,
		Message: "Authentication successful",
	})
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
