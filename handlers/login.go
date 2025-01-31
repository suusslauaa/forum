package handlers

import (
	"forum/database"
	"forum/templates"
	"html/template"
	"net/http"

	"github.com/gofrs/uuid"
)

// LoginRequest структура для логина
type LoginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

var checker string

// LoginHandler обрабатывает запросы на логин
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl, err := template.ParseFS(templates.Files, "login.html")
		if err != nil {
			ErrorHandler(w, "Template loading error", http.StatusInternalServerError)
			return
		}
		data := map[string]interface{}{
			"Check": checker,
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			ErrorHandler(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		if email == "" || password == "" {
			checker = "Email and password are required"
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		db, err := database.InitDB()
		if err != nil {
			ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		user, err := database.GetUserByEmail(db, email)
		if err != nil {
			ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
			return
		}

		if user == nil {
			checker = "Invalid email and password"
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		err = database.ComparePassword(user.PasswordHash, password)
		if err != nil {
			checker = "Invalid email and password"
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		// Успешная аутентификация

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
				MaxAge:   3600,
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
