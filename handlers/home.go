package handlers

import (
	"forum/database"
	"forum/templates"
	"html/template"
	"net/http"
	"strconv"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Простейший in-memory store для сессий
var store = map[string]string{}
var id = map[string]int{}
var (
	// Настройки OAuth2
	oauth2Config = oauth2.Config{
		ClientID:     "",
		ClientSecret: "",
		RedirectURL:  "http://localhost:4000/callback",
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	}
	oauth2StateString = "random"
)

// HomeHandler обрабатывает запросы на главную страницу
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем сессию пользователя
	sessionID, err := GetSessionID(w, r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	// Проверяем, авторизован ли пользователь
	username, loggedIn := store[sessionID.Value]
	UserID := 0
	var role string
	if loggedIn {
		UserID = id[sessionID.Value]

		// Получаем роль пользователя
		db, err := database.InitDB()
		if err != nil {
			ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		role, err = GetUserRole(db, UserID)
		if err != nil {
			ErrorHandler(w, "Error fetching user role", http.StatusInternalServerError)
			return
		}
	}
	Moders := false
	if role == "moder" || role == "admin" {
		Moders = true
	}

	admin := false
	if role == "admin" {
		admin = true
	}

	var categoryID *int
	categoryParam := r.URL.Query().Get("category_id")
	if r.URL.Query().Get("category_id") != "" {
		id, err := strconv.Atoi(categoryParam)
		if err != nil {
			ErrorHandler(w, "Invalid category ID", http.StatusBadRequest)
			return
		}
		categoryID = &id // Передаём указатель, если категория выбрана
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
	tmpl, err := template.ParseFS(templates.Files, "home.html")
	if err != nil {
		ErrorHandler(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"LoggedIn":   loggedIn,
		"ID":         UserID,
		"Username":   username,
		"Posts":      posts,
		"Categories": categories,
		"Moder":      Moders,
		"Admin":      admin,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		ErrorHandler(w, "Template rendering error", http.StatusInternalServerError)
	}
}
