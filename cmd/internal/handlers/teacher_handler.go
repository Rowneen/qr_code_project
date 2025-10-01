package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"qr_code/internal/cookie"
	"qr_code/internal/database"
)

type Lesson struct {
	ID         int    `json:"id"`
	NameLesson string `json:"name_lesson"`
	Date       string `json:"date"`
	TypeLes    string `json:"type_les"`
	QrToken    string `json:"qr_token"`
	IsActive   bool   `json:"is_active"`
	TeacherId  int    `json:"teacher_id"`
}

type TeacherInfoResponse struct {
	Success  bool     `json:"success"`
	Message  string   `json:"message"`
	FullName string   `json:"fullname"`
	Lessons  []Lesson `json:"lessons"`
}

func handler_teacher_getinfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	// 200
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		response := TeacherInfoResponse{
			Success: false,
			Message: "Only GET method allowed",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	// cookie check
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		response := TeacherInfoResponse{
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
		response := TeacherInfoResponse{
			Success: false,
			Message: "Invalid session: " + err.Error(),
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// role check
	if userData["role"] != "Teacher" {
		response := TeacherInfoResponse{
			Success: false,
			Message: "Access denied",
		}
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	db := database.Get()

	// get all teacher lessons
	rows, err := db.Query(`SELECT * FROM lessons WHERE TeacherId = ?`, userData["user_id"])
	if err != nil {
		response := TeacherInfoResponse{
			Success: false,
			Message: "Database error: " + err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer rows.Close()

	var lessons []Lesson

	// read all
	for rows.Next() {
		var lesson Lesson
		err := rows.Scan(&lesson.ID, &lesson.NameLesson, &lesson.Date, &lesson.TypeLes, &lesson.QrToken, &lesson.IsActive, &lesson.TeacherId)
		if err != nil {
			log.Printf("Error scanning lesson: %v\n", err)
			continue
		}
		lessons = append(lessons, lesson)
	}

	// check error iterations
	if err = rows.Err(); err != nil {
		response := TeacherInfoResponse{
			Success: false,
			Message: "Error reading lessons: " + err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	response := TeacherInfoResponse{
		Success:  true,
		Message:  "Lessons retrieved successfully",
		FullName: userData["full_name"].(string),
		Lessons:  lessons,
	}
	json.NewEncoder(w).Encode(response)
}
