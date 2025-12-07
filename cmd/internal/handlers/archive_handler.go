package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"qr_code/internal/cookie"
	"qr_code/internal/cors"
	"qr_code/internal/database"
	"strconv"
)

type ArchiveLesson struct {
	ID         int    `json:"id"`
	NameLesson string `json:"name_lesson"`
	Date       string `json:"date"`
	TypeLes    string `json:"type_les"`
	QrToken    string `json:"qr_token"`
	IsActive   bool   `json:"is_active"`
	TeacherId  int    `json:"teacher_id"`
}

type ArchiveInfoResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Lessons []ArchiveLesson `json:"lessons"`
}

func handler_archive_getlessons(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cors.SetCORSHeaders(&w, r)
	// 200
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		response := ArchiveInfoResponse{
			Success: false,
			Message: "Only GET method allowed",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	// cookie check
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		response := ArchiveInfoResponse{
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
		response := ArchiveInfoResponse{
			Success: false,
			Message: "Invalid session: " + err.Error(),
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// role check
	if userData["role"] != "Teacher" {
		response := ArchiveInfoResponse{
			Success: false,
			Message: "Access denied",
		}
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	db := database.Get()

	// get all teacher lessons
	rows, err := db.Query(`SELECT * FROM lessons WHERE TeacherId = ? AND IsActive = FALSE`, userData["user_id"])
	if err != nil {
		response := ArchiveInfoResponse{
			Success: false,
			Message: "Database error: " + err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer rows.Close()

	var lessons []ArchiveLesson

	// read all
	for rows.Next() {
		var lesson ArchiveLesson
		err := rows.Scan(&lesson.ID, &lesson.NameLesson, &lesson.Date, &lesson.TypeLes, &lesson.QrToken, &lesson.IsActive, &lesson.TeacherId)
		if err != nil {
			log.Printf("Error scanning lesson: %v\n", err)
			continue
		}
		lessons = append(lessons, lesson)
	}

	// check error iterations
	if err = rows.Err(); err != nil {
		response := ArchiveInfoResponse{
			Success: false,
			Message: "Error reading lessons: " + err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	response := ArchiveInfoResponse{
		Success: true,
		Message: "Archive lessons retrieved successfully",
		Lessons: lessons,
	}
	json.NewEncoder(w).Encode(response)
}

func handler_archive_deleteLesson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cors.SetCORSHeaders(&w, r)
	// 200
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "POST" {
		response := ArchiveInfoResponse{
			Success: false,
			Message: "Only POST method allowed",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	// cookie check
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		response := ArchiveInfoResponse{
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
		response := ArchiveInfoResponse{
			Success: false,
			Message: "Invalid session: " + err.Error(),
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// role check
	if userData["role"] != "Teacher" {
		response := ArchiveInfoResponse{
			Success: false,
			Message: "Access denied",
		}
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	lessonIdParam := r.URL.Query().Get("lessonId")
	if lessonIdParam == "" {
		response := ArchiveInfoResponse{
			Success: false,
			Message: "Lesson ID is required",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	// конвертация в int
	lessonId, err := strconv.Atoi(lessonIdParam)
	if err != nil {
		response := ArchiveInfoResponse{
			Success: false,
			Message: "Invalid Lesson ID format",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	if lessonId <= 0 {
		response := ArchiveInfoResponse{
			Success: false,
			Message: "Lesson ID must be positive number",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	db := database.Get()

	// check isactive and owner
	var canDelete int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM lessons WHERE ID = ? AND TeacherId = ? AND IsActive = FALSE`,
		lessonId, userData["user_id"]).Scan(&canDelete)

	if err != nil {
		response := ArchiveInfoResponse{
			Success: false,
			Message: "Database error",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	if canDelete == 0 {
		response := ArchiveInfoResponse{
			Success: false,
			Message: "Lesson not found, you are not the owner, or lesson is still active",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// delete
	_, err = db.Exec(`DELETE FROM lessons WHERE ID = ? AND TeacherId = ?`,
		lessonId, userData["user_id"])

	if err != nil {
		response := ArchiveInfoResponse{
			Success: false,
			Message: "Failed to delete lesson",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ArchiveInfoResponse{
		Success: true,
		Message: "Lesson deleted successfully",
	}
	json.NewEncoder(w).Encode(response)
}
