package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"qr_code/internal/cookie"
	"qr_code/internal/cors"
	"qr_code/internal/database"
	"strconv"
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

type TeacherGetLessonResponse struct {
	Success  bool     `json:"success"`
	Message  string   `json:"message"`
	ID         int    `json:"id"`
	NameLesson string `json:"name_lesson"`
	Date       string `json:"date"`
	TypeLes    string `json:"type_les"`
	QrToken    string `json:"qr_token"`
	IsActive   bool   `json:"is_active"`
	TeacherId  int    `json:"teacher_id"`
}

// ответ
type GetAttendancesResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func handler_teacher_getinfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cors.SetCORSHeaders(&w, r)
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
	rows, err := db.Query(`SELECT * FROM lessons WHERE TeacherId = ? AND IsActive = TRUE`, userData["user_id"])
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

func handler_teacher_getlesson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cors.SetCORSHeaders(&w, r)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		response := TeacherGetLessonResponse{
			Success: false,
			Message: "Only GET method allowed",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	sessionCookie, err := r.Cookie("session")
	if err != nil {
		response := TeacherGetLessonResponse{
			Success: false,
			Message: "Not authorized",
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// valid cookie check
	_, err = cookie.DecryptCookie(sessionCookie.Value)
	if err != nil {
		response := TeacherGetLessonResponse{
			Success: false,
			Message: "Invalid session: " + err.Error(),
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// role check
	/*if userData["role"] != "Teacher" {
		response := TeacherInfoResponse{
			Success: false,
			Message: "Access denied",
		}
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}*/

	// проверка get параметра
	lessonIdParam := r.URL.Query().Get("lessonId")
	if lessonIdParam == "" {
		response := TeacherGetLessonResponse{
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
		response := TeacherGetLessonResponse{
			Success: false,
			Message: "Invalid Lesson ID format",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	if lessonId <= 0 {
		response := TeacherGetLessonResponse{
			Success: false,
			Message: "Lesson ID must be positive number",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	// логика бд
	db := database.Get()

	var lesson TeacherGetLessonResponse
	err = db.QueryRow(`SELECT * FROM lessons WHERE id = ?`, lessonId).Scan(
		&lesson.ID, &lesson.NameLesson, &lesson.Date, &lesson.TypeLes, &lesson.QrToken, &lesson.IsActive, &lesson.TeacherId,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			response := TeacherGetLessonResponse{
				Success: false,
				Message: "Lesson not found",
			}
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(response)
			return
		}
		
		response := TeacherGetLessonResponse{
			Success: false,
			Message: "Database error: " + err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	lesson.Success = true
	lesson.Message = "Lesson retrieved successfully"
	
	json.NewEncoder(w).Encode(lesson)
}

func handler_export_attendances(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cors.SetCORSHeaders(&w, r)

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

	// проверка get параметра
	lessonIdParam := r.URL.Query().Get("lessonId")
	if lessonIdParam == "" {
		response := GetAttendancesResponse{
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
		response := GetAttendancesResponse{
			Success: false,
			Message: "Invalid Lesson ID format",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	if lessonId <= 0 {
		response := GetAttendancesResponse{
			Success: false,
			Message: "Lesson ID must be positive number",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	// логика бд
	db := database.Get()

	// запрос с проверкой айди учителя из куки на соответствие с айди учителя который создал пару в бд
	var teacherID int
	var lessonName string
	err = db.QueryRow(`
		SELECT TeacherId, NameLesson 
		FROM lessons 
		WHERE id = ? AND TeacherId = ?
	`, lessonId, int(userData["user_id"].(float64))).Scan(&teacherID, &lessonName)
	fmt.Printf("[export] teacherID: %d, lessonName: %s\n", teacherID, lessonName)

	if err != nil {
		response := GetAttendancesResponse{
			Success: false,
			Message: "Lesson not found or access denied",
		}
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}
	// запрос на экспорт студентов которые посетили пару
	rows, err := db.Query(`
		SELECT 
			user.FullName,
			user.GroupId,
			attendances.Status,
			attendances.ConfirmedDate
		FROM attendances
		JOIN user ON attendances.StudentId = user.id
		WHERE attendances.LessonId = ?
		ORDER BY user.GroupId, user.FullName
	`, lessonId)

	if err != nil {
		response := GetAttendancesResponse{
			Success: false,
			Message: "Database error: " + err.Error(),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer rows.Close()

	// csv form
	csvContent := "\xEF\xBB\xBF"
	csvContent += "ФИО;Группа;Статус;Дата подтверждения\n"
	// формирование построчно
	for rows.Next() {
		var fullName string
		var groupId int
		var status int
		var confirmedDate sql.NullTime

		err := rows.Scan(&fullName, &groupId, &status, &confirmedDate)
		if err != nil {
			continue
		}
		// говно уберите это
		statusText := "Присутствовал"
		if status == 0 {
			statusText = "Отсутствовал"
		}
		dateText := ""
		if confirmedDate.Valid {
			dateText = confirmedDate.Time.Format("2006-01-02 15:04:05")
		}
		csvContent += fmt.Sprintf("%s;%d;%s;%s\n", fullName, groupId, statusText, dateText)
	}

	// хеадеры для csv файла
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=attendances_%s.csv", lessonName))
	w.Write([]byte(csvContent))
}
