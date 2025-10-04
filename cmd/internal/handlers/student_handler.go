package handlers

import (
	"encoding/json"
	"net/http"
	"qr_code/internal/cookie"
	"qr_code/internal/cors"
)

type StudentInfoResponse struct {
	Success  bool    `json:"success"`
	Message  string  `json:"message"`
	FullName string  `json:"fullname"`
	GroupID  float64 `json:"groupid"`
}

func handler_student_getinfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cors.SetCORSHeaders(&w, r)
	// 200
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != "GET" {
		response := StudentInfoResponse{
			Success: false,
			Message: "Only GET method allowed",
		}
		json.NewEncoder(w).Encode(response)
		return
	}
	// cookie check
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		response := StudentInfoResponse{
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
		response := StudentInfoResponse{
			Success: false,
			Message: "Invalid session: " + err.Error(),
		}
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// role check
	if userData["role"] != "Student" {
		response := StudentInfoResponse{
			Success: false,
			Message: "Access denied",
		}
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := StudentInfoResponse{
		Success:  true,
		Message:  "Profile uploaded successfully",
		FullName: userData["full_name"].(string),
		GroupID:  userData["group_id"].(float64),
	}
	json.NewEncoder(w).Encode(response)
}
