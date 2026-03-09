package routes

import (
	"net/http"

	handlers "github.com/AndreyDubrovin/go_final_project/pkg/api"
	"github.com/go-chi/chi/v5"
)

// SetupRoutes настраивает все роуты приложения
func SetupRoutes(r *chi.Mux, handler *handlers.Handlers) {
	webDir := "web/"

	// API маршруты
	r.Route("/api", func(r chi.Router) {
		r.Get("/nextdate", handler.NextDate.GetNextDate)
		r.Get("/tasks", handler.Task.GetTasks)
		r.Post("/task/done", handler.Task.DoneTask)
		r.Get("/task", handler.Task.GetTask)
		r.Post("/task", handler.Task.AddTask)
		r.Delete("/task", handler.Task.DeleteTask)
		r.Put("/task", handler.Task.PutTask)
	})

	// Статические файлы
	r.Handle("/*", http.FileServer(http.Dir(webDir)))
}
