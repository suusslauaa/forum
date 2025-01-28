package handlers

import (
	"forum/database"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

// CreatePostHandler обрабатывает создание нового поста
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
	}

	_, loggedIn := store[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
	}
	loggedInUserID := id[sessionID.Value]
	// Если метод GET, отображаем форму для создания поста
	if r.Method == http.MethodGet {
		// Получаем список категорий из базы данных
		db, err := database.InitDB()
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		categories, err := database.GetCategories(db)
		if err != nil {
			http.Error(w, "Error fetching categories", http.StatusInternalServerError)
			return
		}

		// Передаем категории в шаблон
		tmpl, err := template.ParseFiles("templates/create_post.html")
		if err != nil {
			http.Error(w, "Template parsing error", http.StatusInternalServerError)
			return
		}

		data := struct {
			UserID     int
			Categories []database.Category
		}{
			UserID:     loggedInUserID, // Получите ID пользователя из сессии
			Categories: categories,
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		userID := id[sessionID.Value]
		categoryID, err := strconv.Atoi(r.FormValue("category"))
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}

		if title == "" || content == "" {
			http.Error(w, "All fields are required", http.StatusBadRequest)
			return
		}

		// Открываем соединение с базой данных
		db, err := database.InitDB()
		if err != nil {
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Вставляем пост в базу данных
		createdAt := time.Now().Format("2006-01-02 15:04:05")
		err = database.CreatePost(db, title, content, userID, categoryID, createdAt)
		if err != nil {
			http.Error(w, "Error saving post to database", http.StatusInternalServerError)
			return
		}

		// Перенаправляем на главную страницу или список постов
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
