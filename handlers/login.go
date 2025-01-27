package handlers

import (
	"forum/database"
	"github.com/gofrs/uuid"
	"html/template"
	"net/http"
)

// LoginRequest структура для логина
type LoginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginHandler обрабатывает запросы на логин
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// Если метод запроса не POST, рендерим HTML-форму для логина
	if r.Method == http.MethodGet {
		// Открываем HTML-шаблон
		tmpl, err := template.ParseFiles("templates/login.html")
		if err != nil {
			http.Error(w, "Template parsing error", http.StatusInternalServerError)
			return
		}

		// Рендерим шаблон
		err = tmpl.Execute(w, nil)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		return
	}

	// Обработка POST-запроса для логина
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		if email == "" || password == "" {
			http.Error(w, "Email and password are required", http.StatusBadRequest)
			return
		}

		db, err := database.InitDB() // Получаем соединение с базой данных
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		user, err := database.GetUserByEmail(db, email) // Получаем пользователя из БД по email
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if user == nil { // Пользователь не найден
			http.Error(w, "Invalid email or password", http.StatusUnauthorized) // 401 Unauthorized
			return
		}

		// Проверяем пароль с помощью bcrypt
		err = database.ComparePassword(user.PasswordHash, password)
		if err != nil { // Пароль не совпадает
			http.Error(w, "Invalid email or password", http.StatusUnauthorized) // 401 Unauthorized
			return
		}

		// Аутентификация успешна! Теперь устанавливаем сессию
		sessionIDCookie, err := r.Cookie("session_id")
		var sessionID string
		if err != nil {
			sessionIDValue := uuid.Must(uuid.NewV4()).String()
			sessionIDCookie = &http.Cookie{
				Name:     "session_id",
				Value:    sessionIDValue,
				HttpOnly: true,
				Secure:   true,
				Path:     "/",
			}
			http.SetCookie(w, sessionIDCookie)
			sessionID = sessionIDValue
		} else {
			sessionID = sessionIDCookie.Value
		}

		store[sessionID] = user.Username[:1]
		id[sessionID] = user.ID

		http.Redirect(w, r, "/", http.StatusFound)
	}
}
