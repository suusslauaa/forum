package handlers

import (
	"forum/database"
	"forum/templates"
	"html/template"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Рендерим HTML-форму для регистрации
		tmpl, err := template.ParseFS(templates.Files, "register.html")
		if err != nil {
			ErrorHandler(w, "Template loading error", http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
		return
	}

	if r.Method == http.MethodPost {
		// Обрабатываем данные из формы
		req := RegisterRequest{
			Email:    r.FormValue("email"),
			Username: r.FormValue("username"),
			Password: r.FormValue("password"),
		}

		// Проверка, существует ли уже email
		emailExists, err := database.CheckEmailExists(req.Email)
		if err != nil {
			checker = "Email verification error"
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}
		if emailExists {
			checker = "Email is already in use"
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		// Хэшируем пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			ErrorHandler(w, "H loading error", http.StatusInternalServerError)
			return
		}

		// Вставка данных в базу
		db, err := database.InitDB()
		if err != nil {
			ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		_, err = db.Exec(
			"INSERT INTO users (email, username, password,role) VALUES (?, ?, ?, 'user')",
			req.Email, req.Username, hashedPassword,
		)
		if err != nil {
			ErrorHandler(w, "Error when creating a user", http.StatusInternalServerError)
			return
		}

		// Перенаправляем на страницу логина после успешной регистрации
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	ErrorHandler(w, "The method is not supported", http.StatusInternalServerError)
}
