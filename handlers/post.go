package handlers

import (
	"fmt"
	"forum/database"
	"github.com/gofrs/uuid"
	"html/template"
	"net/http"
	"strconv"
)

// PostHandler обрабатывает отображение конкретного поста
func PostHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем параметр ID поста из URL
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		// Если сессии нет, создаем новую
		sessionID = &http.Cookie{
			Name:  "session_id",
			Value: uuid.Must(uuid.NewV4()).String(), // Используем gofrs/uuid для генерации UUID
		}
		http.SetCookie(w, sessionID)
	}

	// Проверяем, авторизован ли пользователь
	username, loggedIn := store[sessionID.Value]

	postIDStr := r.URL.Query().Get("id")
	if postIDStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Открываем соединение с базой данных
	db, err := database.InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	// Получаем пост из базы данных
	post, err := database.GetPostByID(db, postID)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}
	fmt.Println(post)

	// Загружаем шаблон и передаем данные
	tmpl, err := template.ParseFiles("templates/post.html")
	if err != nil {
		http.Error(w, "Template parsing error", http.StatusInternalServerError)
		return
	}

	datas := struct {
		LoggedIn bool
		Username string
		Post     database.Post
	}{
		LoggedIn: loggedIn,
		Username: username,
		Post:     post,
	}
	// Прямо вызываем шаблон без дополнительных вызовов WriteHeader
	err = tmpl.Execute(w, datas)
	if err != nil {
		// В случае ошибки рендеринга можно вернуть ошибку
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
