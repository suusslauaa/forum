package handlers

import (
	"forum/database"
	"forum/templates"
	"html/template"
	"log"
	"net/http"
)

type Activity struct {
	ID             int
	AuthorID       int
	Type           string
	PostID         *int
	CommentContent *string
	CreatedAt      string
}

func GetUserActivity(w http.ResponseWriter, r *http.Request) {
	// Проверка сессии
	sessionID, err := GetSessionID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	// _, loggedIn := store[sessionID.Value]

	// Получаем ID пользователя по имени
	userID, _ := id[sessionID.Value]

	query := `
	SELECT a.activity_type, a.created_at, p.id AS post_id, c.content AS comment_content
	FROM activities a
	LEFT JOIN posts p ON a.post_id = p.id
	LEFT JOIN comments c ON a.comment_id = c.id
	WHERE a.user_id = $1
	ORDER BY a.created_at DESC;
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

	var activities []Activity
	for rows.Next() {
		var activity Activity
		// Сканируем в указатели для корректной работы с NULL
		if err := rows.Scan(&activity.Type, &activity.CreatedAt, &activity.PostID, &activity.CommentContent); err != nil {
			log.Println(err)
			continue
		}
		activities = append(activities, activity)
	}

	tmpl, err := template.ParseFS(templates.Files, "activity_page.html")
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
