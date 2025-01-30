package handlers

import (
	"forum/database"
	"forum/templates"
	"html/template"
	"net/http"
)

func UserPostHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем сессию
	sessionID, err := GetSessionID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	// Проверяем, авторизован ли пользователь
	username, loggedIn := store[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Получаем ID пользователя по имени
	UserID, ok := id[sessionID.Value]
	if !ok {
		ErrorHandler(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Открываем соединение с базой данных
	db, err := database.InitDB()
	if err != nil {
		ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	var role string
	role, err = GetUserRole(db, UserID)
	if err != nil {
		ErrorHandler(w, "Error fetching user role", http.StatusInternalServerError)
		return
	}
	Moders := false
	if role == "moder" || role == "admin" {
		Moders = true
	}

	admin := false
	if role == "admin" {
		admin = true
	}
	// Получаем посты пользователя
	posts, err := database.GetPostsByUserID(db, UserID)
	if err != nil {
		ErrorHandler(w, "Error retrieving user's posts", http.StatusInternalServerError)
		return
	}

	// Загружаем шаблон для отображения постов
	tmpl, err := template.ParseFS(templates.Files, "my_posts.html")
	if err != nil {
		ErrorHandler(w, "Template parsing error", http.StatusInternalServerError)
		return
	}

	// Данные для шаблона
	data := map[string]interface{}{
		"LoggedIn": loggedIn,
		"ID":       UserID,
		"Username": username,
		"Posts":    posts,
		"Moder":    Moders,
		"Admin":    admin,
	}

	// Рендерим шаблон
	err = tmpl.Execute(w, data)
	if err != nil {
		ErrorHandler(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
