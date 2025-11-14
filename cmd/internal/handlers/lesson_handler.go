package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"qr_code/internal/cipher"
	"qr_code/internal/cookie"
	"qr_code/internal/cors"
	"qr_code/internal/database"
	"qr_code/internal/utils"
)

// запрос
type LessonRequest struct {
	NameLesson string `json:"name"`
	Date       string `json:"date"`
	TypeLes    string `json:"type"`
	IsActive  bool `json:"isActive"`
	TeacherId int  `json:"teacherId"`
}

// ответ
type LessonResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	QrToken string `json:"qrToken,omitempty"`
}

func handler_lesson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cors.SetCORSHeaders(&w, r)
	// 200
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		response := LessonResponse{
			Success: false,
			Message: "Only POST method allowed",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// cookie check
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		response := LessonResponse{
			Success: false,
			Message: "Not authorized",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// valid cookie check
	userData, err := cookie.DecryptCookie(sessionCookie.Value)
	if err != nil {
		response := LessonResponse{
			Success: false,
			Message: "Invalid session: " + err.Error(),
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// role check
	if userData["role"] != "Teacher" {
		response := LessonResponse{
			Success: false,
			Message: "Access denied",
		}
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("Teacher %s creating lesson", userData["login"])

	var lessonRequest LessonRequest
	decoder := json.NewDecoder(r.Body)
	// проверка на невалидно переданный json (не получается распарсить)
	if err := decoder.Decode(&lessonRequest); err != nil {
		response := LessonResponse{
			Success: false,
			Message: "Invalid JSON format",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// проверка полей на пустоту
	if lessonRequest.Date == "" || lessonRequest.TypeLes == "" || lessonRequest.TeacherId < 0 {
		response := LessonResponse{
			Success: false,
			Message: "Input is empty",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// проверка полей на валидность по символам
	if !utils.IsSafeString(lessonRequest.Date) || !utils.IsSafeString(lessonRequest.TypeLes) {
		response := LessonResponse{
			Success: false,
			Message: "Input contains invalid characters. Only letters, numbers, @, ., -, _ are allowed (3-50 characters)",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	cleanName, cleanDate, cleanTypeLes := utils.CleanString(lessonRequest.NameLesson), utils.CleanString(lessonRequest.Date), utils.CleanString(lessonRequest.TypeLes)

	// база данных логика
	db := database.Get()
	result, err := db.Exec(`
        INSERT INTO lessons (NameLesson, Date, TypeLes, QrToken, IsActive, TeacherId) 
        VALUES (?, ?, ?, ?, ?, ?)`,
		cleanName, cleanDate, cleanTypeLes, "", lessonRequest.IsActive, lessonRequest.TeacherId,
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
	qrToken := generateQrToken(id, cleanName, cleanDate, cleanTypeLes) 

	_, _ = db.Exec(`
		UPDATE lessons 
		SET QrToken = ? 
		WHERE id = ?`,
		qrToken, id,
	)
	response := LessonResponse{
		Success: true,
		Message: fmt.Sprintf("Lesson '%s' created successfully with ID: %d", cleanName, id),
		QrToken: qrToken,
	}
	json.NewEncoder(w).Encode(response)
}

func generateQrToken(id int64, name, date, typeLes string) string {
	data := fmt.Sprintf("%d %s %s %s", id, name, date, typeLes)
	encodedData := cipher.EncodeBase64(data)
	return encodedData 
}

