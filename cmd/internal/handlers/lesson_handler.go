package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"qr_code/internal/cookie"
	"qr_code/internal/cors"
	"qr_code/internal/database"
	"qr_code/internal/qrtoken"
	"qr_code/internal/utils"
)

// запрос создание
type LessonCreateRequest struct {
	NameLesson string `json:"name"`
	Date       string `json:"date"`
	TypeLes    string `json:"type"`
	IsActive   bool   `json:"isActive"`
	TeacherId  int    `json:"teacherId"`
}

// ответ создание
type LessonCreateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	QrToken string `json:"qrToken,omitempty"`
}

// ответ mark
type LessonMarkResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	ID          int64  `json:"id"`
	Name        string `json:"nameLesson"`
	Date        string `json:"date"`
	Type        string `json:"type"`
	TeacherName string `json:"teacherName"`
	Created     int64  `json:"created"`
}

func handler_lessons_create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cors.SetCORSHeaders(&w, r)
	// 200
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		response := LessonCreateResponse{
			Success: false,
			Message: "Only POST method allowed",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// cookie check
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		response := LessonCreateResponse{
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
		response := LessonCreateResponse{
			Success: false,
			Message: "Invalid session: " + err.Error(),
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// role check
	if userData["role"] != "Teacher" {
		response := LessonCreateResponse{
			Success: false,
			Message: "Access denied",
		}
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	var lessonCreateRequest LessonCreateRequest
	decoder := json.NewDecoder(r.Body)
	// проверка на невалидно переданный json (не получается распарсить)
	if err := decoder.Decode(&lessonCreateRequest); err != nil {
		response := LessonCreateResponse{
			Success: false,
			Message: "Invalid JSON format",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	// id in cookie != id in request
	userID, ok := userData["user_id"].(float64)
	if !ok || int(userID) != lessonCreateRequest.TeacherId {
		response := LessonCreateResponse{
			Success: false,
			Message: "Access denied",
		}
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("Teacher %s creating lesson", userData["login"])

	// проверка полей на пустоту
	if lessonCreateRequest.Date == "" || lessonCreateRequest.TypeLes == "" || lessonCreateRequest.TeacherId < 0 {
		response := LessonCreateResponse{
			Success: false,
			Message: "Input is empty",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// проверка полей на валидность по символам
	if !utils.IsSafeString(lessonCreateRequest.Date) || !utils.IsSafeString(lessonCreateRequest.TypeLes) {
		response := LessonCreateResponse{
			Success: false,
			Message: "Input contains invalid characters. Only letters, numbers, @, ., -, _ are allowed (3-50 characters)",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	cleanName, cleanDate, cleanTypeLes := utils.CleanString(lessonCreateRequest.NameLesson), utils.CleanString(lessonCreateRequest.Date), utils.CleanString(lessonCreateRequest.TypeLes)

	// база данных логика
	db := database.Get()
	result, err := db.Exec(`
        INSERT INTO lessons (NameLesson, Date, TypeLes, QrToken, IsActive, TeacherId) 
        VALUES (?, ?, ?, ?, ?, ?)`,
		cleanName, cleanDate, cleanTypeLes, "", lessonCreateRequest.IsActive, lessonCreateRequest.TeacherId,
	)

	if err != nil {
		response := LessonCreateResponse{
			Success: false,
			Message: "Database error: " + err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	id, _ := result.LastInsertId()
	qrToken, _ := qrtoken.Generate(id, cleanName, cleanDate, cleanTypeLes, userData["full_name"].(string))

	_, _ = db.Exec(`
		UPDATE lessons SET QrToken = ? WHERE id = ?`,
		qrToken, id,
	)
	response := LessonCreateResponse{
		Success: true,
		Message: fmt.Sprintf("Lesson '%s' created successfully with ID: %d", cleanName, id),
		QrToken: qrToken,
	}
	json.NewEncoder(w).Encode(response)
}

func handler_lessons_mark(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cors.SetCORSHeaders(&w, r)
	// 200
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" && r.Method != "GET" {
		response := LessonMarkResponse{
			Success: false,
			Message: "Only POST/GET method allowed",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// cookie check
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		response := LessonMarkResponse{
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
		response := LessonMarkResponse{
			Success: false,
			Message: "Invalid session: " + err.Error(),
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// role check
	if userData["role"] != "Student" {
		response := LessonMarkResponse{
			Success: false,
			Message: "Access denied",
		}
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	q := r.URL.Query()
	qrToken := q.Get("token")
	if qrToken == "" {
		response := LessonMarkResponse{
			Success: false,
			Message: "Missing token parameter",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	token, err := qrtoken.Parse(qrToken)
	if err != nil {
		response := LessonMarkResponse{
			Success: false,
			Message: "Invalid or expired QR token",
		}
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	studentID, ok := userData["user_id"].(float64)
	if !ok {
		response := LessonMarkResponse{
			Success: false,
			Message: "Invalid user data",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// db insert
	db := database.Get()

	// check exists
	var count int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM attendances 
		WHERE LessonId = ? AND StudentId = ?`,
		token.ID, int64(studentID),
	).Scan(&count)

	if err != nil {
		response := LessonMarkResponse{
			Success: false,
			Message: "Database error: " + err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// if exists
	if count > 0 {
		response := LessonMarkResponse{
			Success:     false,
			Message: 	 "Attendance already marked",
			ID:          token.ID,
			Name:        token.Name,
			Date:        token.Date,
			Type:        token.Type,
			TeacherName: token.TeacherName,
			Created:     token.Created,
		}
		w.WriteHeader(http.StatusBadRequest) // 400
		json.NewEncoder(w).Encode(response)
		return
	}
	// not exist, add
	_, err = db.Exec(`
		INSERT INTO attendances (LessonId, StudentId, Status, ConfirmedDate) 
		VALUES (?, ?, 1, datetime('now'))`,
		token.ID, int64(studentID),
	)

	if err != nil {
		response := LessonMarkResponse{
			Success: false,
			Message: "Already marked or error",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("ID: %d, Lesson: %s, Teacher: %s\n", token.ID, token.Name, token.TeacherName)
	response := LessonMarkResponse{
		Success:     true,
		Message:     "Attendance marked successfully",
		ID:          token.ID,
		Name:        token.Name,
		Date:        token.Date,
		Type:        token.Type,
		TeacherName: token.TeacherName,
		Created:     token.Created,
	}
	json.NewEncoder(w).Encode(response)
}
