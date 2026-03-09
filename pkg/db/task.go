package db

import (
	"fmt"
	"strconv"
)

// Task описывает задачу для API и слоя работы с базой данных.
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Tasks возвращает список задач, отсортированный по дате.
func Tasks(limit int) ([]*Task, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := DB.Query(
		`SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*Task, 0)
	for rows.Next() {
		var (
			id      int64
			date    string
			title   string
			comment string
			repeat  string
		)
		if err := rows.Scan(&id, &date, &title, &comment, &repeat); err != nil {
			return nil, err
		}
		result = append(result, &Task{
			ID:      strconv.FormatInt(id, 10),
			Date:    date,
			Title:   title,
			Comment: comment,
			Repeat:  repeat,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// GetTask возвращает задачу по идентификатору.
func GetTask(id string) (*Task, error) {
	var (
		intID   int64
		date    string
		title   string
		comment string
		repeat  string
	)

	err := DB.QueryRow(
		`SELECT id, date, title, comment, repeat FROM scheduler WHERE id=?`,
		id,
	).Scan(&intID, &date, &title, &comment, &repeat)
	if err != nil {
		return nil, err
	}

	return &Task{
		ID:      strconv.FormatInt(intID, 10),
		Date:    date,
		Title:   title,
		Comment: comment,
		Repeat:  repeat,
	}, nil
}

// UpdateTask обновляет задачу в базе данных.
func UpdateTask(task *Task) error {
	query := `UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("incorrect id for updating task")
	}
	return nil
}

// DeleteTask удаляет задачу по идентификатору.
func DeleteTask(id string) error {
	res, err := DB.Exec(`DELETE FROM scheduler WHERE id=?`, id)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("incorrect id for deleting task")
	}
	return nil
}

// UpdateDate обновляет только дату задачи по идентификатору.
func UpdateDate(next string, id string) error {
	_, err := DB.Exec(`UPDATE scheduler SET date=? WHERE id=?`, next, id)
	return err
}


