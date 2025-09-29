package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"qr_code/internal/cipher"
	"qr_code/internal/database"
	"qr_code/internal/utils"
)

// запрос
type LessonRequest struct {
	NameLesson string `json:"name"`
	Date       string `json:"date"`
	TypeLes    string `json:"type"`
	// QrToken(само генериться)
	IsActive  bool `json:"isActive"`
	TeacherId int  `json:"teacherId"`
}

// ответ
type LessonResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	QrToken string `json:"qrToken,omitempty"` // добавляем поле для возврата QR токена
}

func handler_lesson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		response := AuthResponse{
			Success: false,
			Message: "Only POST method allowed",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	var lessonRequest LessonRequest
	decoder := json.NewDecoder(r.Body)
	// проверка на невалидно переданный json (не получается распарсить)
	if err := decoder.Decode(&lessonRequest); err != nil {
		response := AuthResponse{
			Success: false,
			Message: "Invalid JSON format",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// проверка полей на пустоту
	if lessonRequest.Date == "" || lessonRequest.TypeLes == "" || lessonRequest.TeacherId < 0 {
		response := AuthResponse{
			Success: false,
			Message: "Input is empty",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// проверка полей на валидность по символам
	if !utils.IsSafeString(lessonRequest.Date) || !utils.IsSafeString(lessonRequest.TypeLes) {
		response := AuthResponse{
			Success: false,
			Message: "Input contains invalid characters. Only letters, numbers, @, ., -, _ are allowed (3-50 characters)",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	cleanName, cleanDate, cleanTypeLes := utils.CleanString(lessonRequest.NameLesson), utils.CleanString(lessonRequest.Date), utils.CleanString(lessonRequest.TypeLes)
	qrToken := generateQrToken(cleanName, cleanDate, cleanTypeLes)

	// база данных логика
	db := database.Get()
	result, err := db.Exec(`
        INSERT INTO lessons (NameLesson, Date, TypeLes, QtToken, IsActive, TeacherId) 
        VALUES (?, ?, ?, ?, ?, ?)`,
		cleanName, cleanDate, cleanTypeLes, qrToken, lessonRequest.IsActive, lessonRequest.TeacherId,
	)

	if err != nil {
		response := LessonResponse{
			Success: false,
			Message: "Database error: " + err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Получаем ID вставленной записи
	id, _ := result.LastInsertId()

	response := LessonResponse{
		Success: true,
		Message: fmt.Sprintf("Lesson '%s' created successfully with ID: %d", cleanName, id),
		QrToken: qrToken,
	}
	json.NewEncoder(w).Encode(response)
}

func generateQrToken(name, date, typeLes string) string {
	data := fmt.Sprintf("%s %s %s", name, date, typeLes)
	encodedData := cipher.EncodeBase64(data)
	return encodedData
}
