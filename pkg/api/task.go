package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"final_project/pkg/db"
)

// taskHandler обрабатывает GET, POST, PUT и DELETE запросы к /api/task.
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleTaskGet(w, r)
	case http.MethodPost:
		handleTaskPost(w, r)
	case http.MethodPut:
		handleTaskPut(w, r)
	case http.MethodDelete:
		handleTaskDelete(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// handleTaskGet возвращает параметры задачи по идентификатору.
func handleTaskGet(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeJSONError(w, "Identifier not specified")
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeJSONError(w, "Task not found")
		return
	}

	writeJSON(w, task)
}

// handleTaskPost добавляет новую задачу.
func handleTaskPost(w http.ResponseWriter, r *http.Request) {
	var t db.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeJSONError(w, err.Error())
		return
	}

	// Валидация заголовка.
	if strings.TrimSpace(t.Title) == "" {
		writeJSONError(w, "Не указан заголовок задачи")
		return
	}

	now := time.Now()

	if err := normalizeTask(now, &t); err != nil {
		writeJSONError(w, err.Error())
		return
	}

	id, err := db.AddTask(&t)
	if err != nil {
		writeJSONError(w, err.Error())
		return
	}

	writeJSON(w, map[string]any{
		"id": id,
	})
}

// handleTaskPut обновляет задачу по данным из JSON.
func handleTaskPut(w http.ResponseWriter, r *http.Request) {
	var t db.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeJSONError(w, err.Error())
		return
	}

	t.ID = strings.TrimSpace(t.ID)
	if t.ID == "" {
		writeJSONError(w, "Identifier not specified")
		return
	}

	// Валидация заголовка.
	if strings.TrimSpace(t.Title) == "" {
		writeJSONError(w, "Не указан заголовок задачи")
		return
	}

	now := time.Now()

	if err := normalizeTask(now, &t); err != nil {
		writeJSONError(w, err.Error())
		return
	}

	if err := db.UpdateTask(&t); err != nil {
		writeJSONError(w, err.Error())
		return
	}

	// Успешное обновление — пустой JSON объект.
	writeJSON(w, map[string]any{})
}

// normalizeTask нормализует дату и правило повторения задачи.
func normalizeTask(now time.Time, t *db.Task) error {
	const layout = "20060102"
	today := now.Format(layout)

	// Обработка даты.
	t.Date = strings.TrimSpace(t.Date)
	if t.Date == "" {
		t.Date = today
	} else {
		d, err := time.Parse(layout, t.Date)
		if err != nil {
			return fmt.Errorf("invalid date format")
		}

		if d.Before(time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())) {
			// Дата в прошлом.
			if strings.TrimSpace(t.Repeat) == "" {
				t.Date = today
			} else {
				next, err := NextDate(now, t.Date, t.Repeat)
				if err != nil {
					return err
				}
				t.Date = next
			}
		}
	}

	// Валидация правила repeat.
	if strings.TrimSpace(t.Repeat) != "" {
		if _, err := NextDate(now, t.Date, t.Repeat); err != nil {
			return err
		}
	}

	return nil
}

// handleTaskDelete удаляет задачу по идентификатору.
func handleTaskDelete(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeJSONError(w, "Identifier not specified")
		return
	}

	if err := db.DeleteTask(id); err != nil {
		writeJSONError(w, err.Error())
		return
	}

	writeJSON(w, map[string]any{})
}


