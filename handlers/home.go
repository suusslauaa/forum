package handlers

import (
	"context"
	"encoding/json"
	"forum/database"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid" // Используем новый пакет для UUID
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

var goauth2Config = oauth2.Config{
	ClientID:     "",                     // Замените на ваш Client ID
	ClientSecret: "", // Замените на ваш Client Secret
	RedirectURL:  "http://localhost:4000/github/callback",    // Убедитесь, что ваш redirect URL совпадает с настройками приложения на GitHub
	Scopes:       []string{"user:email"},                     // GitHub предоставляет доступ к email и данным пользователя
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	},
}

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	// Генерация URL для редиректа на Google OAuth
	url := oauth2Config.AuthCodeURL(oauth2StateString, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

func GitHubLogin(w http.ResponseWriter, r *http.Request) {
	url := goauth2Config.AuthCodeURL(oauth2StateString, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusFound)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	// Получаем код, который Google вернул
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	// Обмен кода на токен
	token, err := oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Unable to get token", http.StatusInternalServerError)
		return
	}

	// Используем токен для запроса информации о пользователе
	client := oauth2Config.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		http.Error(w, "Unable to fetch user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Декодируем информацию о пользователе
	var userInfo struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Unable to parse user info", http.StatusInternalServerError)
		return
	}

	// Логируем информацию о пользователе
	log.Printf("Google User: %s, Email: %s", userInfo.Name, userInfo.Email)

	// Открываем соединение с базой данных
	db, err := database.InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Проверяем, существует ли пользователь в базе данных
	user, err := database.GetUserByEmail(db, userInfo.Email)
	if err != nil {
		http.Error(w, "Error checking user", http.StatusInternalServerError)
		return
	}

	// Если пользователя нет, создаем его в базе данных
	if user == nil {
		err = database.CreateUser(db, userInfo.Email, userInfo.Name, "defaultpassword")
		if err != nil {
			http.Error(w, "Error creating user", http.StatusInternalServerError)
			return
		}
		// Получаем созданного пользователя
		user, err = database.GetUserByEmail(db, userInfo.Email)
		if err != nil {
			http.Error(w, "Error fetching user after creation", http.StatusInternalServerError)
			return
		}
	}

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

	// Сохраняем ID пользователя в store, используя sessionID
	store[sessionID.Value] = userInfo.Name[:1]
	id[sessionID.Value] = user.ID

	// Создаем cookie для сессии
	sessionIDCookie := &http.Cookie{
		Name:     "session_id",
		Value:    sessionID.Value,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	}
	http.SetCookie(w, sessionIDCookie)
	log.Printf("User info: %+v", userInfo)
	log.Printf("Session ID: %s", sessionID.Value)

	// Перенаправляем на главную страницу
	http.Redirect(w, r, "/", http.StatusFound)
}

func GitHubCallback(w http.ResponseWriter, r *http.Request) {
	// Получаем код от GitHub
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Code not found", http.StatusBadRequest)
		return
	}

	// Обмен кодом на токен
	token, err := goauth2Config.Exchange(context.Background(), code)
	if err != nil {
		http.Error(w, "Unable to get token", http.StatusInternalServerError)
		return
	}

	// Создаем клиент для доступа к API GitHub
	client := goauth2Config.Client(context.Background(), token)

	// Получаем информацию о пользователе
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		http.Error(w, "Unable to fetch user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Декодируем информацию о пользователе
	var userInfo struct {
		Login string `json:"login"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Unable to parse user info", http.StatusInternalServerError)
		return
	}

	// Логируем информацию о пользователе
	log.Printf("GitHub User: %s, Email: %s", userInfo.Login, userInfo.Email)

	// Открываем соединение с базой данных
	db, err := database.InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Проверяем, существует ли пользователь в базе данных
	user, err := database.GetUserByEmail(db, userInfo.Email)
	if err != nil {
		http.Error(w, "Error checking user", http.StatusInternalServerError)
		return
	}

	// Если пользователя нет, создаем его в базе данных
	if user == nil {
		err = database.CreateUser(db, userInfo.Email, userInfo.Login, "defaultpassword")
		if err != nil {
			http.Error(w, "Error creating user", http.StatusInternalServerError)
			return
		}
		// Получаем созданного пользователя
		user, err = database.GetUserByEmail(db, userInfo.Email)
		if err != nil {
			http.Error(w, "Error fetching user after creation", http.StatusInternalServerError)
			return
		}
	}

	// Создаем сессию для пользователя
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		sessionID = &http.Cookie{
			Name:  "session_id",
			Value: uuid.Must(uuid.NewV4()).String(),
		}
		http.SetCookie(w, sessionID)
	}

	// Сохраняем ID пользователя в store
	store[sessionID.Value] = userInfo.Login[:1]
	id[sessionID.Value] = user.ID

	// Перенаправляем на главную страницу
	http.Redirect(w, r, "/", http.StatusFound)
}

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
