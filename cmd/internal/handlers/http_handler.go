package handlers

import (
	"log"
	"net/http"
	"qr_code/internal/config"
)

// регистрация хандлеров и запуск
func RegisterHTTPHandlers() {
	cfg := config.Get()
	// запуск обработчиков на функции
	http.HandleFunc("/auth", handler_auth)
	// lessons
	http.HandleFunc("/lessons/create", handler_lessons_create)
	http.HandleFunc("/lessons/mark", handler_lessons_mark)
	// teacher
	http.HandleFunc("/teacher/getInfo", handler_teacher_getinfo)
	http.HandleFunc("/teacher/export", handler_export_attendances)
	// student
	http.HandleFunc("/student/getInfo", handler_student_getinfo)
	// logout
	http.HandleFunc("/logout", LogoutHandler)
	// конфиг сервера
	httpServer := &http.Server{
		Handler:      nil,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	// HTTPS сервер
	go func() {
		httpsServer := &http.Server{
			Handler:      nil,
			ReadTimeout:  cfg.HTTPServer.Timeout,
			WriteTimeout: cfg.HTTPServer.Timeout,
			IdleTimeout:  cfg.HTTPServer.IdleTimeout,
		}
		httpsServer.Addr = cfg.HTTPServer.HttpsAddress
		log.Printf("Starting HTTPS server on %s\n", httpsServer.Addr)
		err := httpsServer.ListenAndServeTLS("cert/server.crt", "cert/server.key")
		if err != nil {
			log.Printf("HTTPS server error: %v", err)
		}
	}()

	// HTTP сервер основной
	httpServer.Addr = cfg.HTTPServer.HttpAddress
	log.Printf("Starting HTTP server on %s\n", httpServer.Addr)
	httpServer.ListenAndServe()
}
