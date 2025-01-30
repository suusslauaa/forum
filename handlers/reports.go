package handlers

import (
	"database/sql"
	"forum/database" // Импортируй правильный путь к database
	"forum/templates"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

// Структура для хранения данных репорта
type Report struct {
	ID        int
	PostID    int
	PostTitle string
	Reporter  string
}

// ReportsHandler - обработчик для отображения открытых репортов
func ReportsHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем сессию пользователя
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	_, loggedIn := store[sessionID.Value]
	UserID := id[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Подключаемся к базе данных
	db, err := database.InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Println("Ошибка подключения к БД:", err)
		return
	}
	defer db.Close()

	// Проверяем роль пользователя (только admin может видеть страницу)
	role, err := GetUserRole(db, UserID)
	if err != nil {
		http.Error(w, "Error fetching user role", http.StatusInternalServerError)
		return
	}
	if role != "admin" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// SQL-запрос для получения открытых репортов
	rows, err := db.Query(`
		SELECT reports.id, reports.post_id, COALESCE(posts.title, 'Deleted Post'), users.username
		FROM reports
		LEFT JOIN posts ON reports.post_id = posts.id
		LEFT JOIN users ON reports.reported_by = users.id
		WHERE reports.status = 'open'
	`)
	if err != nil {
		http.Error(w, "Database query error", http.StatusInternalServerError)
		log.Println("Ошибка запроса к БД:", err)
		return
	}
	defer rows.Close()

	// Создаём массив для хранения репортов
	var reports []Report
	for rows.Next() {
		var report Report
		var postTitle sql.NullString

		// Сканируем данные, но обрабатываем NULL для названия поста
		if err := rows.Scan(&report.ID, &report.PostID, &postTitle, &report.Reporter); err != nil {
			http.Error(w, "Error scanning reports", http.StatusInternalServerError)
			log.Println("Ошибка при сканировании:", err)
			return
		}

		// Если `title` == NULL, заменяем на "Deleted Post"
		if postTitle.Valid {
			report.PostTitle = postTitle.String
		} else {
			report.PostTitle = "Deleted Post"
		}

		reports = append(reports, report)
	}

	// Проверяем ошибки после `rows.Next()`
	if err = rows.Err(); err != nil {
		http.Error(w, "Database iteration error", http.StatusInternalServerError)
		log.Println("Ошибка обработки строк:", err)
		return
	}

	if r.Method == http.MethodPost {
		action := r.URL.Query().Get("action")
		reportID := r.URL.Query().Get("id") // Получаем ID репорта из запроса

		if reportID == "" {
			ErrorHandler(w, "Missing report ID", http.StatusBadRequest)
			return
		}
		reportIDInt, _ := strconv.Atoi(reportID)
		switch action {
		case "ignore":
			database.DeletePostReport(db, reportIDInt)
		default:
			ErrorHandler(w, "Invalid action", http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/home", http.StatusFound)
	}

	// Загружаем HTML-шаблон
	tmpl, err := template.ParseFS(templates.Files, "reports.html")
	if err != nil {
		http.Error(w, "Template parsing error", http.StatusInternalServerError)
		log.Println("Ошибка загрузки шаблона:", err)
		return
	}

	// Отправляем данные в шаблон
	err = tmpl.Execute(w, reports)
	if err != nil {
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
		log.Println("Ошибка рендеринга:", err)
	}
}
