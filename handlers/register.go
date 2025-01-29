package handlers

import (
	"forum/database"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"net/http"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Рендерим HTML-форму для регистрации
		tmpl, err := template.ParseFiles("templates/register.html")
		if err != nil {
			http.Error(w, "Ошибка загрузки шаблона", http.StatusInternalServerError)
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
			http.Error(w, "Ошибка при хэшировании пароля", http.StatusInternalServerError)
			return
		}

		// Вставка данных в базу
		db, err := database.InitDB()
		if err != nil {
			http.Error(w, "Ошибка подключения к базе данных", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		_, err = db.Exec(
			"INSERT INTO users (email, username, password) VALUES (?, ?, ?)",
			req.Email, req.Username, hashedPassword,
		)
		if err != nil {
			http.Error(w, "Ошибка при создании пользователя", http.StatusInternalServerError)
			return
		}

		// Перенаправляем на страницу логина после успешной регистрации
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
}
