package handlers

import (
	"log"

	"github.com/AndreyDubrovin/go_final_project/pkg/api/nextdate"
	"github.com/AndreyDubrovin/go_final_project/pkg/api/task"
	"github.com/AndreyDubrovin/go_final_project/pkg/repository"
)

// Handlers содержит все обработчики приложения
type Handlers struct {
	NextDate *nextdate.NextDateHandler
	Task     *task.TaskHandler
	logger   *log.Logger
}

// NewHandler создаёт новый набор обработчиков
func NewHandler(TaskRepo repository.TaskRepo, logger *log.Logger) *Handlers {
	// Создаем обработчики
	NextDateHandler := nextdate.NextDateHendler(logger)
	TaskHandler := task.TaskHendler(TaskRepo, logger)
	return &Handlers{
		NextDate: NextDateHandler,
		Task:     TaskHandler,
		logger:   logger,
	}
}
