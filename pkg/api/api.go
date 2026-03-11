package api

import (
	"net/http"
	"os"
)

var envPwd string

// Init регистрирует HTTP-маршруты API.
func Init() {
	envPwd = os.Getenv("TODO_PASSWORD")
	http.HandleFunc("/api/tasks", authMiddleware(tasksHandler))
	http.HandleFunc("/api/task", authMiddleware(taskHandler))
	http.HandleFunc("/api/task/done", authMiddleware(taskDoneHandler))
	http.HandleFunc("/api/nextdate", nextDateHandler)
	http.HandleFunc("/api/signin", signInHandler)
}
