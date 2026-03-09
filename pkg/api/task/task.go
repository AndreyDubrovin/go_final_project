package task

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/AndreyDubrovin/go_final_project/pkg/api/nextdate"
	"github.com/AndreyDubrovin/go_final_project/pkg/models"
	"github.com/AndreyDubrovin/go_final_project/pkg/repository"
)

type TaskHandler struct {
	taskRepo repository.TaskRepo
	logger   *log.Logger
}

func TaskHendler(taskRepo repository.TaskRepo, logger *log.Logger) *TaskHandler {
	return &TaskHandler{
		taskRepo: taskRepo,
		logger:   logger,
	}
}

var Limit = 50

func (h *TaskHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})
	search := r.FormValue("search")
	if search != "" {
		tasks, err := h.taskRepo.SearchTasks(search, Limit)
		if err != nil {
			sendErrorMessage(w, "Ошибка получения задач", http.StatusInternalServerError, h.logger)
			h.logger.Printf("Ошибка получения задач: %v", err)
			return
		}
		if len(*tasks) == 0 {
			response["tasks"] = make([]int, 0)
		} else {
			response["tasks"] = *tasks
		}
	} else {
		tasks, err := h.taskRepo.GetTasks(Limit)
		if err != nil {
			sendErrorMessage(w, "Ошибка получения задач", http.StatusInternalServerError, h.logger)
			h.logger.Printf("Ошибка получения задач: %v", err)
			return
		}
		if len(*tasks) == 0 {
			response["tasks"] = make([]int, 0)
		} else {
			response["tasks"] = *tasks
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Ошибка кодирования JSON: %v", err)
	}
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	var response *models.Task
	id := r.FormValue("id")
	if id != "" {
		task, err := h.taskRepo.GetTask(id)
		if err != nil {
			sendError(w, err, http.StatusInternalServerError, h.logger)
			h.logger.Printf("Ошибка получения задач: %v", err)
			return
		}
		response = task
	} else {
		sendErrorMessage(w, "Не указан идентификатор", http.StatusInternalServerError, h.logger)
		h.logger.Printf("Не указан идентификатор")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Ошибка кодирования JSON: %v", err)
	}
}
func (h *TaskHandler) AddTask(w http.ResponseWriter, r *http.Request) {
	var req models.Task
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	today := time.Now()
	now := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	// Валидация
	if req.Title == "" {
		sendErrorMessage(w, "Не указан заголовок задачи", http.StatusBadRequest, h.logger)
		return
	}
	if req.Date == "" {
		req.Date = now.Format(nextdate.DateFormat)
	}
	if req.Date != "" {
		t, err := time.Parse(nextdate.DateFormat, req.Date)
		if err != nil {
			sendError(w, err, http.StatusBadRequest, h.logger)
			return
		}
		var nextDateFind string
		if len(req.Repeat) != 0 {
			nextDateFind, err = nextdate.NextDate(now.Format(nextdate.DateFormat), req.Date, req.Repeat)
			if err != nil {

				h.logger.Printf("Ошибка в расчётах NextDate: %v", err)
				sendError(w, err, http.StatusBadRequest, h.logger)
				return
			}
		}
		if nextdate.AfterNow(now, t) {
			// если NOW больше req.Data:
			if len(req.Repeat) == 0 {
				// если правила повторения нет, то берём сегодняшнее число
				req.Date = now.Format(nextdate.DateFormat)
			} else {
				// если повторения есть, то делаем рассчет, какую дату нужно поставить в календать. Так как указан req.Data в прошлом.
				req.Date = nextDateFind
			}
		}
	}

	// Вставляем в базу
	taskId, err := h.taskRepo.AddTask(&req)
	if err != nil {
		sendErrorMessage(w, "Ошибка при создании задачи", http.StatusInternalServerError, h.logger)
		return
	}
	// Успешный ответ
	response := map[string]interface{}{
		"id": taskId,
	}

	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Ошибка кодирования JSON: %v", err)
	}
}

func (h *TaskHandler) DoneTask(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	response := map[string]interface{}{}
	id := r.FormValue("id")
	if id == "" {
		sendErrorMessage(w, "Не указан идентификатор", http.StatusBadRequest, h.logger)
		h.logger.Printf("Не указан идентификатор")
		return
	}
	task, err := h.taskRepo.GetTask(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || strings.Contains(err.Error(), "не найдена") {
			sendError(w, err, http.StatusNotFound, h.logger) // 404 для не найденных задач
			return
		} else {
			sendError(w, err, http.StatusInternalServerError, h.logger) // 500 для других ошибок
			h.logger.Printf("Ошибка получения задач: %v", err)
			return
		}

	}
	// если повторения нет, то удалить задачу.
	if len(task.Repeat) == 0 {
		if err := h.taskRepo.DeleteTask(id); err != nil {
			sendError(w, err, http.StatusInternalServerError, h.logger)
			h.logger.Printf("Ошибка удаления задачи: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			h.logger.Printf("Ошибка кодирования JSON: %v", err)
		}
		return
	}
	// если есть повторения, вычисляем новую дату, меняем в базе.
	nextDate, err := nextdate.NextDate(now.Format(nextdate.DateFormat), task.Date, task.Repeat)
	if err != nil {
		h.logger.Printf("Ошибка в NextDate: %v", err)
		sendError(w, err, http.StatusBadRequest, h.logger)
		return
	}
	if err := h.taskRepo.PutTaskDate(task.Id, nextDate); err != nil {
		sendError(w, err, http.StatusInternalServerError, h.logger)
		h.logger.Printf("Ошибка измененения даты в задаче: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Ошибка кодирования JSON: %v", err)
	}
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{}
	id := r.FormValue("id")
	if id == "" {
		sendErrorMessage(w, "Не указан идентификатор", http.StatusInternalServerError, h.logger)
		h.logger.Printf("Не указан идентификатор")
		return
	}
	if err := h.taskRepo.DeleteTask(id); err != nil {
		sendError(w, err, http.StatusInternalServerError, h.logger)
		h.logger.Printf("Ошибка удаления задачи: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Ошибка кодирования JSON: %v", err)
	}
}

func (h *TaskHandler) PutTask(w http.ResponseWriter, r *http.Request) {
	var req models.Task
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	today := time.Now()
	now := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	// Валидация
	if req.Title == "" {
		sendErrorMessage(w, "Не указан заголовок задачи", http.StatusBadRequest, h.logger)
		return
	}
	if req.Date == "" {
		req.Date = now.Format(nextdate.DateFormat)
	}
	if req.Date != "" {
		t, err := time.Parse(nextdate.DateFormat, req.Date)
		if err != nil {
			sendError(w, err, http.StatusBadRequest, h.logger)
			return
		}
		var nextDateFind string
		if len(req.Repeat) != 0 {
			nextDateFind, err = nextdate.NextDate(now.Format(nextdate.DateFormat), req.Date, req.Repeat)
			if err != nil {
				h.logger.Printf("Ошибка в расчётах NextDate: %v", err)
				sendError(w, err, http.StatusBadRequest, h.logger)
				return
			}
		}
		if nextdate.AfterNow(now, t) {
			// если NOW больше req.Data:
			if len(req.Repeat) == 0 {
				// если правила повторения нет, то берём сегодняшнее число
				req.Date = now.Format(nextdate.DateFormat)
			} else {
				// если повторения есть, то делаем рассчет, какую дату нужно поставить в календать. Так как указан req.Data в прошлом.
				req.Date = nextDateFind
			}
		}

	}
	// обновляем документ
	if err := h.taskRepo.PutTask(&req); err != nil {
		sendError(w, err, http.StatusInternalServerError, h.logger)
		return
	}
	response := map[string]interface{}{}
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Ошибка кодирования JSON: %v", err)
	}
}

func sendErrorMessage(w http.ResponseWriter, message string, statusCode int, logger *log.Logger) {
	response := map[string]interface{}{
		"error": message,
	}
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Printf("Ошибка кодирования JSON: %v", err)
	}
}
func sendError(w http.ResponseWriter, message error, statusCode int, logger *log.Logger) {
	response := map[string]interface{}{
		"error": message.Error(),
	}
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Printf("Ошибка кодирования JSON: %v", err)
	}
}
