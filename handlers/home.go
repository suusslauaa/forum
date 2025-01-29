package handlers

import (
	"forum/database"
	"github.com/gofrs/uuid" // Используем новый пакет для UUID
	"html/template"
	"net/http"
	"strconv"
)

// Простейший in-memory store для сессий
var store = map[string]string{}
var id = map[string]int{}

// HomeHandler обрабатывает запросы на главную страницу
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем сессию пользователя
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
	UserID := 0
	if loggedIn {
		UserID = id[sessionID.Value]
	}

	categoryID := 0
	if r.URL.Query().Get("category_id") != "" {
		categoryID, err = strconv.Atoi(r.URL.Query().Get("category_id"))
		if err != nil {
			ErrorHandler(w, "Invalid category ID", http.StatusBadRequest)
			return
		}
	}

	// Получаем посты из базы данных
	db, err := database.InitDB()
	if err != nil {
		ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	posts, err := database.GetPosts(db, categoryID)
	if err != nil {
		ErrorHandler(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}

	// Получаем категории из базы данных
	categories, err := database.GetCategories(db)
	if err != nil {
		ErrorHandler(w, "Error fetching categories", http.StatusInternalServerError)
		return
	}

	// Загружаем шаблон и передаем данные
	tmpl, err := template.ParseFiles("./templates/home.html")
	if err != nil {
		ErrorHandler(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Создаем данные для шаблона, включая информацию о пользователе
	data := struct {
		LoggedIn   bool
		ID         int
		Username   string
		Posts      []database.Post
		Categories []database.Category
	}{
		LoggedIn:   loggedIn,
		ID:         UserID,
		Username:   username,
		Posts:      posts,
		Categories: categories,
	}

	tmpl.Execute(w, data)
}
