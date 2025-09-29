package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"qr_code/internal/cipher"
	"qr_code/internal/database"
	"qr_code/internal/utils"
)

// запрос
type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// ответ
type AuthResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	FullName string `json:"fullname"`
	Role     string `json:"role"`
	GroupId  int    `json:"groupid"`
}

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
	if !utils.IsSafeString(authRequest.Login) || !utils.IsSafeString(authRequest.Password) {
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
	// очистка паролей от спецсимволов
	cleanLogin, cleanPassword := utils.CleanString(authRequest.Login), utils.CleanString(authRequest.Password)

	var (
		id       int
		Login    string
		PassHash string
		FullName string
		Role     string
		GroupId  sql.NullInt64
	)
	// запрос к бд
	err := db.QueryRow(
		"SELECT * FROM user WHERE Login = ? AND PassHash = ? LIMIT 1",
		cleanLogin, cipher.MD5(cleanPassword),
	).Scan(&id, &Login, &PassHash, &FullName, &Role, &GroupId)

	// проверка на ошибку после запроса
	if err != nil {
		// log.Printf("db error: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(AuthResponse{
			Success: false,
			Message: "Invalid login or password",
		})
		return
	}
	if Role == "Student" {
		// если сюда прошло - запрос корректно прошел
		log.Printf("correct auth: %d, %s, %s, %s, %s, %d", id, Login, PassHash, FullName, Role, GroupId.Int64)
		json.NewEncoder(w).Encode(AuthResponse{
			Success:  true,
			Message:  "Authentication successful",
			FullName: FullName,
			Role:     Role,
			GroupId:  int(GroupId.Int64),
		})

	} else {
		// если сюда прошло - запрос корректно прошел
		log.Printf("correct auth: %d, %s, %s, %s, %s, %d", id, Login, PassHash, FullName, Role, GroupId.Int64)
		json.NewEncoder(w).Encode(AuthResponse{
			Success:  true,
			Message:  "Authentication successful",
			FullName: FullName,
			Role:     Role,
		})

	}

}
