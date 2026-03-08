package nextdate

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

type NextDateHandler struct {
	logger *log.Logger
}

func NextDateHendler(logger *log.Logger) *NextDateHandler {
	return &NextDateHandler{
		logger: logger,
	}
}

func (h *NextDateHandler) GetNextDate(w http.ResponseWriter, r *http.Request) {
	now := r.FormValue("now")
	date := r.FormValue("date")
	r.ParseForm()
	// Проверяем наличие параметра
	if _, exists := r.Form["repeat"]; exists {
		// Параметр существует (даже если значение пустое)
		repeat := r.FormValue("repeat")
		if repeat == "" {
			h.logger.Printf("параметр repeat пустой")
			http.Error(w, "параметр repeat пустой", http.StatusBadRequest)
			return
		}
		// параметр есть, работаем дальше
		result, err := NextDate(now, date, repeat)
		if err != nil {
			h.logger.Printf("Ошибка в NextDate: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write([]byte(result))
	} else {
		http.Error(w, "параметр repeat отсутствует в запросе", http.StatusBadRequest)
		return
	}

}

func NextDate(now, date, repeat string) (string, error) {
	// текущая дата
	parsedNow, err := time.Parse("20060102", now)
	if err != nil {
		return "", fmt.Errorf("ошибка парсинга now: %w", err)
	}
	// установленная дата задачи
	parsedDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("ошибка парсинга date: %w", err)
	}
	// срез интервалов
	intervals := strings.Split(repeat, " ")
	if len(intervals) == 0 {
		return "", fmt.Errorf("Интервал не задан")
	}
	switch intervals[0] {
	case "d":
		resultDate, err := repeatD(intervals, parsedNow, parsedDate)
		return resultDate, err
	case "y":
		resultDate, err := repeatY(intervals, parsedNow, parsedDate)
		return resultDate, err
	case "w":
		resultDate, err := repeatW(intervals, parsedNow, parsedDate)
		return resultDate, err
	case "m":
		resultDate, err := repeatM(intervals, parsedNow, parsedDate)
		return resultDate, err
	default:
		return "", fmt.Errorf("ошибка в формате: %s", intervals[0])
	}
}

func repeatD(intervals []string, parsedNow, parsedDate time.Time) (string, error) {
	if len(intervals) < 2 {
		return "", fmt.Errorf("для правила d требуется указать интервал")
	}
	num, err := strconv.Atoi(intervals[1])
	if err != nil {
		return "", fmt.Errorf("ошибка преобразования интервала в число: %w", err)
	}
	if num < 0 {
		return "", fmt.Errorf("интервал не может быть отрицательным")
	}
	if num > 400 {
		return "", fmt.Errorf("превышен максимально допустимый интервал (400)")
	}
	for {
		parsedDate = parsedDate.AddDate(0, 0, num)
		if AfterNow(parsedDate, parsedNow) {
			break
		}
	}
	return parsedDate.Format("20060102"), nil
}

func repeatY(intervals []string, parsedNow, parsedDate time.Time) (string, error) {
	if len(intervals) > 1 {
		return "", fmt.Errorf("для правила y не требуется указывать интервал")
	}
	for {
		parsedDate = parsedDate.AddDate(1, 0, 0)
		if AfterNow(parsedDate, parsedNow) {
			break
		}
	}
	return parsedDate.Format("20060102"), nil
}

func repeatW(intervals []string, parsedNow, parsedDate time.Time) (string, error) {
	// заполненные интервалы
	if len(intervals) > 2 {
		return "", fmt.Errorf("для правила W некорректно указан интервал")
	}
	if len(intervals) < 2 {
		return "", fmt.Errorf("для правила W не указаны дни недели")
	}
	weeksIntervals := strings.Split(intervals[1], ",")
	// мапа для дней недели
	weeksNumber := map[time.Weekday]int{
		time.Monday:    1,
		time.Tuesday:   2,
		time.Wednesday: 3,
		time.Thursday:  4,
		time.Friday:    5,
		time.Saturday:  6,
		time.Sunday:    7,
	}
	dates := make([]time.Time, 0, 7)
	// если дата меньше текущей, то отталкиваеся от текущий и ищем ближайшую следующую дату.
	for _, str := range weeksIntervals {
		checkDate := parsedNow
		if AfterNow(parsedDate, parsedNow) { // если установленная дата больше текущий то ищем следущую дату от установленной
			checkDate = parsedDate
		}
		number, err := strconv.Atoi(str)
		if err != nil {
			return "", fmt.Errorf("ошибка преобразования интервала в число: %w", err)
		}
		// если число больше 7 или меньше 1 - это ошибка
		if number < 1 || number > 7 {
			return "", fmt.Errorf("недопустимое значение в дне недели: %d", number)
		}
		for {
			checkDate = checkDate.AddDate(0, 0, 1)
			weekdayNumberForResultDate := checkDate.Weekday()
			numberOfWeekForResultDate := weeksNumber[weekdayNumberForResultDate]
			if numberOfWeekForResultDate == number {
				break
			}
		}
		// добавляем подходящую дату в слайс
		dates = append(dates, checkDate)
	}
	resultDate := dates[0]
	// ищем ближайшую дату в слайсе
	for _, date := range dates[1:] {
		if date.Before(resultDate) {
			resultDate = date
		}
	}
	return resultDate.Format("20060102"), nil
}

func repeatM(intervals []string, parsedNow, parsedDate time.Time) (string, error) {
	switch {
	case len(intervals) < 2 || len(intervals) > 3:
		return "", fmt.Errorf("для правила m некорректно указан интервал")
	case len(intervals) == 2: // если интервала 2 то учитываем только ближайшие числа, последний и предпоследний день.
		result, err := checkTwoConditions(intervals, parsedNow, parsedDate)
		if err != nil {
			return "", err
		}
		return result, nil
	case len(intervals) == 3: // если интервала 3 то учитываем месяца и дни
		result, err := checkThreeConditions(intervals, parsedNow, parsedDate)
		if err != nil {
			return "", err
		}
		return result, nil
	}
	return "", nil
}

func checkTwoConditions(intervals []string, parsedNow, parsedDate time.Time) (string, error) {
	dates := make([]time.Time, 0, 33) // 31 день + последний и предпоследний день
	dateNumbersIntervals := strings.Split(intervals[1], ",")
	// если дата меньше текущей, то отталкиваеся от текущий и ищем ближайшую следующую дату.
	for _, str := range dateNumbersIntervals {
		number, err := strconv.Atoi(str)
		if err != nil {
			return "", fmt.Errorf("ошибка преобразования интервала в число: %w", err)
		}
		checkDate := parsedNow
		if AfterNow(parsedDate, parsedNow) { // если установленная дата больше текущий то ищем следущую дату от установленной
			checkDate = parsedDate
		}
		// если число больше 31 или меньше -2 - это ошибка выход за пределы.
		if number < -2 || number > 31 {
			return "", fmt.Errorf("недопустимое значение: %d", number)
		}
		// отдельно обрабатываем -1 и -2
		switch number {
		case -1:
			isLastDay := isLastDayOfMonth(checkDate)
			if isLastDay {
				// Если это последний день, переходим к последнему дню следующего месяца
				findDate := getLastDayOfMonth(checkDate.AddDate(0, 1, 0))
				dates = append(dates, findDate)
			} else {
				// Если не последний день, устанавливаем последний день текущего месяца
				findDate := getLastDayOfMonth(checkDate)
				dates = append(dates, findDate)
			}
		case -2:
			// Проверяем, является ли текущий день предпоследним днем месяца
			isPrevLastDay := isPrevLastDayOfMonth(checkDate)
			if isPrevLastDay {
				// Если это предпоследний день, переходим к предпоследнему дню следующего месяца
				findDate := getPrevLastDayOfMonth(checkDate.AddDate(0, 1, 0))
				dates = append(dates, findDate)
			} else {
				// Если не предпоследний день, устанавливаем предпоследний день текущего месяца
				findDate := getPrevLastDayOfMonth(checkDate)
				dates = append(dates, findDate)
			}
		default:
			// перебираем дни и ищем ближайший день к выбранной дате.
			for {
				checkDate = checkDate.AddDate(0, 0, 1)
				numberOfDay := checkDate.Day()
				if numberOfDay == number {
					break
				}
			}
			// добавляем подходящую дату в слайс
			dates = append(dates, checkDate)
		}
	}
	resultDate := dates[0]
	// ищем ближайшую дату в слайсе
	for _, date := range dates[1:] {
		if date.Before(resultDate) {
			resultDate = date
		}
	}
	return resultDate.Format("20060102"), nil
}

func checkThreeConditions(intervals []string, parsedNow, parsedDate time.Time) (string, error) {
	months := make([]int, 0, 12) // 12 месяцев сохраняем месяцы, которые больше checkDate
	days := make([]int, 0, 12)   // числа дней в запросе
	dateNumbersIntervals := strings.Split(intervals[1], ",")
	monthNumbersIntervals := strings.Split(intervals[2], ",")
	fmt.Println("dateNumbersIntervals:", dateNumbersIntervals, " monthNumbersIntervals:", monthNumbersIntervals)
	// если дата меньше текущей, то отталкиваеся от текущий и ищем ближайшую следующую дату.
	checkDate := parsedNow
	if AfterNow(parsedDate, parsedNow) { // если установленная дата больше текущий то используем установленную
		checkDate = parsedDate
	}
	//текущее число месяца
	checkDateMonthInt := int(checkDate.Month())
	// добавляем месяца, которые больше checkDateMonthInt в слайс
	for _, str := range monthNumbersIntervals {
		number, err := strconv.Atoi(str)
		if err != nil {
			return "", fmt.Errorf("ошибка преобразования интервала в число: %w", err)
		}
		if number > checkDateMonthInt {
			months = append(months, number)
		}
	}
	// добавляем дни в слайс
	for _, str := range dateNumbersIntervals {
		number, err := strconv.Atoi(str)
		if err != nil {
			return "", fmt.Errorf("ошибка преобразования интервала в число: %w", err)
		}
		days = append(days, number)
	}
	if len(months) == 0 {
		for _, str := range monthNumbersIntervals {
			number, err := strconv.Atoi(str)
			if err != nil {
				return "", fmt.Errorf("ошибка преобразования интервала в число: %w", err)
			}
			months = append(months, number)
		}
		sort.Ints(months)
		sort.Ints(days)
		newDate := time.Date(
			checkDate.Year()+1,
			time.Month(months[0]),
			days[0],
			0, 0, 0, 0,
			time.UTC,
		)
		return newDate.Format("20060102"), nil
	}
	sort.Ints(months)
	sort.Ints(days)
	// если есть месяца, которые больше текущего месяца, подставляем этот месяц и день.
	if len(months) > 0 {
		newDate := time.Date(
			checkDate.Year(),
			time.Month(months[0]),
			days[0],
			0, 0, 0, 0,
			time.UTC,
		)
		return newDate.Format("20060102"), nil
	}
	return "", nil
}

// Проверяет, является ли день последним в месяце
func isLastDayOfMonth(t time.Time) bool {
	// Добавляем один день и проверяем, изменился ли месяц
	nextDay := t.AddDate(0, 0, 1)
	return nextDay.Month() != t.Month()
}

// Возвращает последний день месяца для указанной даты
func getLastDayOfMonth(t time.Time) time.Time {
	// Переходим к первому дню следующего месяца и вычитаем один день
	// Это дает последний день текущего месяца
	firstDayNextMonth := time.Date(
		t.Year(),
		t.Month()+1,
		1,
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Nanosecond(),
		t.Location(),
	)
	return firstDayNextMonth.AddDate(0, 0, -1)
}

func isPrevLastDayOfMonth(t time.Time) bool {
	// Добавляем два дня и проверяем, изменился ли месяц
	// Если через два дня месяц меняется, значит сегодня предпоследний день
	twoDaysLater := t.AddDate(0, 0, 2)
	return twoDaysLater.Month() != t.Month()
}

// Возвращает предпоследний день месяца для указанной даты
func getPrevLastDayOfMonth(t time.Time) time.Time {
	// Получаем последний день месяца
	lastDay := getLastDayOfMonth(t)
	// Вычитаем один день для предпоследнего
	return lastDay.AddDate(0, 0, -1)
}

// Проверка на большую дату
func AfterNow(date, now time.Time) bool {
	return date.After(now)
}
