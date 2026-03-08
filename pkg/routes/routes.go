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
	r.Handle("/css/*", http.StripPrefix("/css/", http.FileServer(http.Dir(webDir+"css"))))
	r.Handle("/js/*", http.StripPrefix("/js/", http.FileServer(http.Dir(webDir+"js"))))

	// HTML страницы
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, webDir+"index.html")
	})
	r.Get("/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, webDir+"index.html")
	})
	r.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, webDir+"login.html")
	})
	r.Get("/login.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, webDir+"login.html")
	})
	// Favicon
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, webDir+"favicon.ico")
	})

}
