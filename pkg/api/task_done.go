package api

import (
	"net/http"
	"strings"
	"time"

	"final_project/pkg/db"
)

// taskDoneHandler обрабатывает POST /api/task/done.
func taskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

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

	// Если задача не повторяющаяся — удаляем её.
	if strings.TrimSpace(task.Repeat) == "" {
		if err := db.DeleteTask(id); err != nil {
			writeJSONError(w, err.Error())
			return
		}
		writeJSON(w, map[string]any{})
		return
	}

	// Периодическая задача — вычисляем следующую дату.
	now := time.Now()
	next, err := NextDate(now, task.Date, task.Repeat)
	if err != nil {
		writeJSONError(w, err.Error())
		return
	}

	if err := db.UpdateDate(next, id); err != nil {
		writeJSONError(w, err.Error())
		return
	}

	writeJSON(w, map[string]any{})
}

