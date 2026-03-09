package db

import (
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
)

// DB — глобальное соединение с базой данных (устанавливается в Init).
var DB *sql.DB

// schema содержит SQL-команды для создания таблицы scheduler
// и индекса по колонке date.
const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date VARCHAR NOT NULL,
    title VARCHAR NOT NULL,
    comment TEXT,
    repeat VARCHAR
);

CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);
`

// Init открывает базу данных по пути dbFile и при необходимости создаёт таблицу и индекс.
func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)

	DB, err = sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}

	if install {
		_, err = DB.Exec(schema)
		if err != nil {
			_ = DB.Close()
			return err
		}
	}

	return nil
}

// AddTask добавляет задачу в таблицу scheduler и возвращает её идентификатор.
func AddTask(task *Task) (int64, error) {
	res, err := DB.Exec(
		`INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`,
		task.Date, task.Title, task.Comment, task.Repeat,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

