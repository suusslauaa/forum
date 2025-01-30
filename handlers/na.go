package handlers

import (
	"database/sql"
	"forum/database"
	"forum/templates"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

func NotificationsHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := GetSessionID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	username, loggedIn := store[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	// Получаем ID пользователя по имени
	userID, _ := id[sessionID.Value]

	query := `
	SELECT n.id, n.notification_type, n.created_at, p.id AS post_id, c.content AS comment_content, n.is_read
	FROM notifications n
	LEFT JOIN posts p ON n.post_id = p.id
	LEFT JOIN comments c ON n.comment_id = c.id
	WHERE n.user_id = $1
	ORDER BY n.created_at DESC;
	`
	// Открываем соединение с базой данных
	db, err := database.InitDB()
	if err != nil {
		ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	if r.Method == http.MethodPost && loggedIn {
		// Получаем ID поста из параметров URL
		postIDStr := r.URL.Query().Get("id")
		if postIDStr == "" {
			ErrorHandler(w, "Post ID is required", http.StatusBadRequest)
			return
		}

		ID, err := strconv.Atoi(postIDStr)

		if err != nil {
			ErrorHandler(w, "Invalid post ID", http.StatusBadRequest)
			return
		}
		if err := handleNotificationActions(w, r, db, ID); err != nil {
			ErrorHandler(w, "Error handling post action", http.StatusInternalServerError)
			return
		}

		// Перенаправляем на страницу поста с добавленным комментарием
		http.Redirect(w, r, "/notifications", http.StatusSeeOther)
		return
	}
	rows, err := db.Query(query, userID)
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var activities []Notification
	for rows.Next() {
		var activity Notification
		// Сканируем в указатели для корректной работы с NULL
		if err := rows.Scan(&activity.ID, &activity.Type, &activity.CreatedAt, &activity.PostID, &activity.CommentContent, &activity.IsRead); err != nil {
			log.Println(err)
			continue
		}
		activities = append(activities, activity)
	}

	tmpl, err := template.ParseFiles("templates/notifications.html")
	if err != nil {
		log.Println("Error loading template:", err)
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	// Передаем данные в шаблон
	data := map[string]interface{}{
		"Activities": activities,
		"Username":   username,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

type Notification struct {
	ID             int
	Type           string
	PostID         *int
	CommentContent *string
	CreatedAt      string
	IsRead         bool
}

func handleNotificationActions(w http.ResponseWriter, r *http.Request, db *sql.DB, ID int) error {
	action := r.URL.Query().Get("action")
	switch action {
	case "read":
		database.ReadNotification(db, ID)
		http.Redirect(w, r, "/notifications", http.StatusSeeOther)
	default:
		ErrorHandler(w, "Invalid action", http.StatusBadRequest)
		return nil
	}

	return nil
}

func ReadNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := GetSessionID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
	username, loggedIn := store[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	// Получаем ID пользователя по имени
	userID, _ := id[sessionID.Value]

	query := `
	SELECT n.id, n.notification_type, n.created_at, p.id AS post_id, c.content AS comment_content, n.is_read
	FROM notifications n
	LEFT JOIN posts p ON n.post_id = p.id
	LEFT JOIN comments c ON n.comment_id = c.id
	WHERE n.user_id = $1
	ORDER BY n.created_at DESC;
	`
	// Открываем соединение с базой данных
	db, err := database.InitDB()
	if err != nil {
		ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query(query, userID)
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var activities []Notification
	for rows.Next() {
		var activity Notification
		// Сканируем в указатели для корректной работы с NULL
		if err := rows.Scan(&activity.ID, &activity.Type, &activity.CreatedAt, &activity.PostID, &activity.CommentContent, &activity.IsRead); err != nil {
			log.Println(err)
			continue
		}
		activities = append(activities, activity)
	}

	tmpl, err := template.ParseFS(templates.Files, "readnotifications.html")
	if err != nil {
		log.Println("Error loading template:", err)
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	// Передаем данные в шаблон
	data := map[string]interface{}{
		"Activities": activities,
		"Username":   username,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}
