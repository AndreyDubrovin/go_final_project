package server

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	handlers "github.com/AndreyDubrovin/go_final_project/pkg/api"
	"github.com/AndreyDubrovin/go_final_project/pkg/repository"
	"github.com/AndreyDubrovin/go_final_project/pkg/routes"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Logger     *log.Logger
	HTTPServer *http.Server
}

func NewServer(db *sql.DB, serverPort string, logger *log.Logger) *Server {
	// подгружаем обработчики
	handler := handlers.NewHandler(
		repository.NewTaskRepository(db),
		logger,
	)
	// Настраиваем роуты
	r := chi.NewRouter()

	// Все роуты
	routes.SetupRoutes(r, handler)

	server := &Server{
		Logger: logger,
		HTTPServer: &http.Server{
			Addr:         serverPort,
			Handler:      r,
			ErrorLog:     logger,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second,
		},
	}

	return server
}
