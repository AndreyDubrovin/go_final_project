// package db предоставляет функционал для работы с базой данных
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite" // Импорт драйвера
)

// LoadEnv загружает переменные из .env
func LoadEnv() error {
	if err := godotenv.Load(); err != nil {
		return err
	}
	return nil
}

// ConnectToDB устанавливает соединение с БД
func ConnectToDB(logger *log.Logger) (*sql.DB, error) {
	dbFile := os.Getenv("TODO_DBFILE")
	_, err := os.Stat(dbFile)

	var newDb bool
	if err != nil {
		newDb = true
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		fmt.Println("Ошибка при открытии базы данных:", err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	if newDb == true {
		err := CreateTableInDb(db, logger)
		if err != nil {
			return nil, err
		}
	}
	return db, nil
}

func CreateTableInDb(db *sql.DB, logger *log.Logger) error {
	schema := `	
			CREATE TABLE scheduler (
    		id 			INTEGER PRIMARY KEY AUTOINCREMENT,
    		date 		CHAR(8) NOT NULL DEFAULT '',
    		title 		VARCHAR(256) NOT NULL DEFAULT '',
    		comment 	TEXT NOT NULL DEFAULT '',
    		repeat 		VARCHAR(128) NOT NULL DEFAULT ''
			);
			CREATE INDEX idx_scheduler_date ON scheduler(date);
		`
	_, err := db.Exec(schema)
	if err != nil {
		logger.Printf("ошибка создания таблицы: %v", err)
		return fmt.Errorf("ошибка создания таблицы: %w", err)
	}
	fmt.Println("Таблица scheduler успешно создана")
	return nil
}
