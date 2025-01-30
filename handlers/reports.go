package handlers

import (
	"database/sql"
	"forum/database" // Импортируй правильный путь к database
	"forum/templates"
	"html/template"
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
	sessionID, err := GetSessionID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
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
		ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Проверяем роль пользователя (только admin может видеть страницу)
	role, err := GetUserRole(db, UserID)
	if err != nil {
		ErrorHandler(w, "Error fetching user role", http.StatusInternalServerError)
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
			ErrorHandler(w, "Error scanning reports", http.StatusInternalServerError)
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
		ErrorHandler(w, "Database iteration error", http.StatusInternalServerError)
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
		ErrorHandler(w, "Template parsing error", http.StatusInternalServerError)
		return
	}
	Moders := true
	admin := true
	data := map[string]interface{}{
		"Moder":    Moders,
		"Admin":    admin,
		"UserRole": role,
	}
	// Отправляем данные в шаблон
	err = tmpl.Execute(w, data)
	if err != nil {
		ErrorHandler(w, "Template rendering error", http.StatusInternalServerError)
	}
}
