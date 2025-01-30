package handlers

import (
	"forum/database"
	"html/template"
	"log"
	"net/http"
)

func NotificationsHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка сессии
	sessionID, err := getSessionID(w, r)
	if err != nil {
		ErrorHandler(w, "Session error", http.StatusInternalServerError)
		return
	}

	// _, loggedIn := store[sessionID.Value]

	// Получаем ID пользователя по имени
	userID, _ := id[sessionID.Value]

	query := `
	SELECT n.notification_type, n.created_at, p.id AS post_id, c.content AS comment_content, n.is_read
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
		if err := rows.Scan(&activity.Type, &activity.CreatedAt, &activity.PostID, &activity.CommentContent, &activity.IsRead); err != nil {
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
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

type Notification struct {
	Type           string
	PostID         *int
	CommentContent *string
	CreatedAt      string
	IsRead         string
}
