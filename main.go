package main

import (
	"log"
	"os"

	"github.com/AndreyDubrovin/go_final_project/pkg/db"
	"github.com/AndreyDubrovin/go_final_project/pkg/server"
)

func main() {
	// Загружаем переменные из .env
	if err := db.LoadEnv(); err != nil {
		log.Fatal("Ошибка загрузки .env:", err)
	}
	log.Println("Переменные окружения загружены")

	// Запускаем Логи
	flog, err := os.OpenFile(`server.log`, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer flog.Close()
	logger := log.New(flog, "SERVER: ", log.LstdFlags|log.Lshortfile)

	// Подключаемся к БД
	db, err := db.ConnectToDB(logger)
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}
	defer db.Close()
	log.Println("Успешно подключились к sql!")

	// Запускаем сервер
	serverPort := os.Getenv("TODO_PORT")
	if serverPort == "" {
		serverPort = ":7540"
	}
	// создаём сервер
	s := server.NewServer(db, serverPort, logger)
	// запускаем сервер
	if err := s.HTTPServer.ListenAndServe(); err != nil {
		logger.Fatal(err)
	}
}
