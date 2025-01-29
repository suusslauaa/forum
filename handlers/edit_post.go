package handlers

import (
	"forum/database"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

func EditPostHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
	}

	_, loggedIn := store[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
	}
	loggedInUserID := id[sessionID.Value]

	postIDStr := r.URL.Query().Get("id")
	if postIDStr == "" {
		ErrorHandler(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		ErrorHandler(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Если метод GET, отображаем форму для создания поста
	if r.Method == http.MethodGet {
		// Получаем список категорий из базы данных
		db, err := database.InitDB()
		if err != nil {
			ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Получаем пост из базы данных
		post, err := database.GetPostByID(db, postID)
		if err != nil {
			ErrorHandler(w, "Post not found", http.StatusNotFound)
			return
		}

		categories, err := database.GetCategories(db)
		if err != nil {
			ErrorHandler(w, "Error fetching categories", http.StatusInternalServerError)
			return
		}

		// Передаем категории в шаблон
		tmpl, err := template.ParseFiles("templates/edit_post.html")
		if err != nil {
			ErrorHandler(w, "Template parsing error", http.StatusInternalServerError)
			return
		}
		data := struct {
			Post       database.Post
			UserID     int
			Categories []database.Category
			Check      string
		}{
			UserID:     loggedInUserID, // Получите ID пользователя из сессии
			Categories: categories,
			Post:       post,
			Check:      checker,
		}
		if checker != "" {
			checker = ""
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			ErrorHandler(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		userID := r.FormValue("user_id")
		categoryID, err := strconv.Atoi(r.FormValue("category"))
		if err != nil {
			ErrorHandler(w, "Invalid category ID", http.StatusBadRequest)
			return
		}

		if title == "" || content == "" || userID == "" {
			checker = "All fields are required"
			http.Redirect(w, r, "/edit-post?$="+strconv.Itoa(postID), http.StatusSeeOther)
			return
		}

		// Открываем соединение с базой данных
		db, err := database.InitDB()
		if err != nil {
			ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Вставляем пост в базу данных
		createdAt := time.Now().Format("2006-01-02 15:04:05")
		err = database.EditPost(db, title, content, categoryID, createdAt, postID)
		if err != nil {
			ErrorHandler(w, "Error saving post to database", http.StatusInternalServerError)
			return
		}

		// Перенаправляем на главную страницу или список постов
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
