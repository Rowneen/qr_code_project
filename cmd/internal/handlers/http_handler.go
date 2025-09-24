package handlers

import (
	"io"
	"net/http"
)

// сами хандлеры обработчики
func handler_auth(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Write([]byte("only post"))
		return
	}
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	w.Write([]byte(string(body)))
}

// регистрация хандлеров и запуск
func RegisterHandlers() {
	http.HandleFunc("/auth", handler_auth)
	http.ListenAndServe(":8080", nil)
}
