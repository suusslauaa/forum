package handlers

import (
	"forum/database"
	"forum/templates"
	"html/template"
	"net/http"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Рендеринг формы
		tmpl, err := template.ParseFS(templates.Files, "register.html")
		if err != nil {
			ErrorHandler(w, "Template loading error", http.StatusInternalServerError)
			return
		}
		data := struct{ Check string }{Check: checker}
		if checker != "" {
			checker = ""
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			ErrorHandler(w, "Template rendering error", http.StatusInternalServerError)
		}
		return
	}

	if r.Method == http.MethodPost {
		req := RegisterRequest{
			Email:    strings.ToLower(strings.TrimSpace(r.FormValue("email"))),
			Username: strings.TrimSpace(r.FormValue("username")),
			Password: r.FormValue("password"),
		}
		// Валидация полей
		if err := validateCredentials(req); err != nil {
			checker = err.Error()
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		// Проверка уникальности email и username
		if exists, err := database.CheckEmailExists(req.Email); err != nil || exists {
			checker = "Email is already in use"
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		if exists, err := database.CheckUsernameExists(req.Username); err != nil || exists {
			checker = "The username is already taken"
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		// Хэширование пароля
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			ErrorHandler(w, "Password hashing error", http.StatusInternalServerError)
			return
		}

		// Сохранение в БД
		db, err := database.InitDB()
		if err != nil {
			ErrorHandler(w, "Error connecting to the database", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		if _, err = db.Exec(
			"INSERT INTO users (email, username, password, role) VALUES (?, ?, ?, 'user')",
			req.Email, req.Username, hashedPassword,
		); err != nil {
			ErrorHandler(w, "User creation error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ErrorHandler(w, "The method is not supported", http.StatusMethodNotAllowed)
}

// Валидация входных данных
