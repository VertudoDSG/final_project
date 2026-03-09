package api

import "net/http"

// Init регистрирует HTTP-маршруты API.
func Init() {
	http.HandleFunc("/api/tasks", authMiddleware(tasksHandler))
	http.HandleFunc("/api/task", authMiddleware(taskHandler))
	http.HandleFunc("/api/task/done", authMiddleware(taskDoneHandler))
	http.HandleFunc("/api/nextdate", nextDateHandler)
	http.HandleFunc("/api/signin", signInHandler)
}

