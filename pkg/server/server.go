package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"final_project/pkg/api"
)

const defaultPort = 7540

// Run запускает HTTP-сервер приложения.
func Run() error {
	port := defaultPort

	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 && p < 65536 {
			port = p
		}
	}

	// Корень приложения: отдаём index.html из ./web
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			http.ServeFile(w, r, filepath.Join("web", "index.html"))
			return
		}

		// Все остальные пути — статические файлы из ./web
		http.StripPrefix("/", http.FileServer(http.Dir("web"))).ServeHTTP(w, r)
	})

	// Регистрация API-маршрутов.
	api.Init()

	addr := ":" + strconv.Itoa(port)
	log.Printf("Starting server on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

