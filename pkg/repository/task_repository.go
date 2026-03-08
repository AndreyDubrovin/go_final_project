package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AndreyDubrovin/go_final_project/pkg/models"
)

type TaskRepo interface {
	GetTasks(limit int) (*[]models.Task, error)
	GetTask(id string) (*models.Task, error)
	AddTask(access *models.Task) (int64, error)
	SearchTasks(text string, limit int) (*[]models.Task, error)
	PutTask(task *models.Task) error
	DeleteTask(id string) error
	PutTaskDate(id, date string) error
}

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db: db}
}
func (r *TaskRepository) SearchTasks(text string, limit int) (*[]models.Task, error) {
	var rows *sql.Rows
	var err error
	query := ""
	isDate := isDate(text)
	if isDate {
		t, err := time.Parse("02.01.2006", text)
		if err != nil {
			return nil, err
		}
		resultDate := t.Format("20060102")
		fmt.Println("result", resultDate)
		query = `
        SELECT id, date, title, comment, repeat 
        FROM scheduler 
		WHERE date = ?
        ORDER BY date ASC 
        LIMIT ?
    `
		rows, err = r.db.Query(query, resultDate, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	} else {
		query = `
        SELECT id, date, title, comment, repeat 
        FROM scheduler 
		WHERE LOWER(title) LIKE LOWER(?) OR LOWER(comment) LIKE LOWER(?)
        ORDER BY date ASC 
        LIMIT ?
    `
		searchPattern := "%" + text + "%"
		rows, err = r.db.Query(query, searchPattern, searchPattern, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
	}

	// Слайс для хранения результатов
	var tasks []models.Task

	for rows.Next() {
		var r models.Task
		err := rows.Scan(&r.Id, &r.Date, &r.Title, &r.Comment, &r.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &tasks, nil
}

func (r *TaskRepository) GetTasks(limit int) (*[]models.Task, error) {
	rows, err := r.db.Query(`
        SELECT id, date, title, comment, repeat 
        FROM scheduler 
        ORDER BY date ASC 
        LIMIT ?
    `, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Слайс для хранения результатов
	var tasks []models.Task

	// Итерируем по результатам запроса
	for rows.Next() {
		var r models.Task
		err := rows.Scan(&r.Id, &r.Date, &r.Title, &r.Comment, &r.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, r)
	}
	// Проверяем на ошибки после итерации
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return &tasks, nil
}
func (r *TaskRepository) GetTask(id string) (*models.Task, error) {
	query := `
        SELECT id, date, title, comment, repeat 
        FROM scheduler 
        WHERE id = ?
    `
	var task models.Task
	err := r.db.QueryRow(query, id).Scan(
		&task.Id,
		&task.Date,
		&task.Title,
		&task.Comment,
		&task.Repeat,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("задача с ID %s не найдена", id)
		}
		return nil, fmt.Errorf("ошибка при поиске задачи: %w", err)
	}

	return &task, nil
}

func (r *TaskRepository) AddTask(task *models.Task) (int64, error) {
	var id int64
	query := `INSERT INTO scheduler (date, title, comment, repeat) 
              VALUES (?, ?, ?, ?)`

	res, err := r.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err == nil {
		id, err = res.LastInsertId()
	}
	return id, err
}

func (r *TaskRepository) DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("ошибка при удалении задачи: %w", err)
	}
	// Проверяем, была ли удалена запись
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка при получении количества удаленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("задача с ID %s не найдена", id)
	}

	return nil
}
func (r *TaskRepository) PutTask(task *models.Task) error {
	query := `
        UPDATE scheduler 
        SET date = ?, title = ?, comment = ?, repeat = ?
        WHERE id = ?
    `
	res, err := r.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.Id)
	if err != nil {
		return err
	}
	// метод RowsAffected() возвращает количество записей к которым
	// была применена SQL команда
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}

func (r *TaskRepository) PutTaskDate(id, date string) error {
	query := `
        UPDATE scheduler 
        SET date = ?
        WHERE id = ?
    `
	res, err := r.db.Exec(query, date, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`некорректный id`)
	}
	return nil
}
func isDate(text string) bool {
	parsed, err := time.Parse("02.01.2006", text)
	if err != nil {
		return false
	}
	// Проверяем, что после парсинга форматирование дает ту же строку
	return parsed.Format("02.01.2006") == text
}
