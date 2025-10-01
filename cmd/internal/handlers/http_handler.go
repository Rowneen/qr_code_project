package handlers

import (
	"net/http"
	"qr_code/internal/config"
)

// сами хандлеры обработчики
// ...

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
	http.HandleFunc("/lessons", handler_lesson)
	http.HandleFunc("/teacher/getInfo", handler_teacher_getinfo)
	http.HandleFunc("/student/getInfo", handler_student_getinfo)
	server.ListenAndServe()
}
