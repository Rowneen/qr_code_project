package handlers

import (
	"log"
	"net/http"
	"qr_code/internal/config"
)

// сами хандлеры обработчики
// ...

// регистрация хандлеров и запуск
func RegisterHTTPHandlers() {
	cfg := config.Get()
	// запуск обработчиков на функции
	http.HandleFunc("/auth", handler_auth)
	http.HandleFunc("/lessons", handler_lesson)
	http.HandleFunc("/teacher/getInfo", handler_teacher_getinfo)
	http.HandleFunc("/student/getInfo", handler_student_getinfo)
	http.HandleFunc("/logout", LogoutHandler)
	// конфиг сервера
	server := &http.Server{
		Handler:      nil,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	go func() {
		server.Addr = cfg.HTTPServer.HttpsAddress
		log.Printf("Starting HTTPS server on %s\n", server.Addr)
		err := server.ListenAndServeTLS("cert/server.crt", "cert/server.key")
		if err != nil {
			log.Printf("HTTPS server error: %v", err)
		}
	}()

	// HTTP сервер основной
	server.Addr = cfg.HTTPServer.HttpAddress
	log.Printf("Starting HTTP server on %s\n", server.Addr)
	server.ListenAndServe()
}
