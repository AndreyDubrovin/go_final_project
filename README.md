# Файлы для итогового задания

В директории `tests` находятся тесты для проверки API, которое должно быть реализовано в веб-сервере.

Директория `web` содержит файлы фронтенда.

Аутенфикация и работа с докером не выполнялась

В .env следующие настройки:
TODO_PORT=:7540
TODO_DBFILE=scheduler.db

в tests/settings.go настройки:
var Port = 7540
var DBFile = "../scheduler.db"
var FullNextDate = true
var Search = true
var Token = ``

## Структура проекта

go_final_project/
├── main.go - основой файл
├── pkg/ - Основные пакеты приложения
│ ├── api/
│ │ ├── nextdate/ - Пакет для расчёта следующую дату repeat
│ │ └── task/ - Пакет для работы с api (CURD)
│ ├── db/ - Пакет для подключения к DB
│ ├── models/ - Структуры для работы с DB
│ ├── repository/ - Обработчики запросов с api
│ ├── routes/ - роуты
│ ├── server/ - настройка сервера
├── pkg/ - Тесты
