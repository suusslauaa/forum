package handlers

import (
	"forum/database"
	"forum/templates"
	"html/template"
	"log"
	"net/http"
)

type User struct {
	ID       int
	Username string
	Email    string
	Role     string
}

func UserListHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем сессию пользователя
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	_, loggedIn := store[sessionID.Value]
	UserID := id[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Подключаемся к базе данных
	db, err := database.InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Println("Ошибка подключения к БД:", err)
		return
	}
	defer db.Close()

	// Проверяем роль пользователя (только admin может видеть страницу)
	role, err := GetUserRole(db, UserID)
	if err != nil {
		http.Error(w, "Error fetching user role", http.StatusInternalServerError)
		return
	}
	if role != "admin" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Если это POST-запрос - выполняем изменение роли
	if r.Method == http.MethodPost {
		userID := r.URL.Query().Get("id")
		action := r.URL.Query().Get("action")

		switch action {
		case "promote":
			_, err := db.Exec(`UPDATE users SET role = 'moderator' WHERE id = ? AND role = 'user'`, userID)
			if err != nil {
				http.Error(w, "Error promoting user", http.StatusInternalServerError)
				return
			}
		case "demote":
			_, err := db.Exec(`UPDATE users SET role = 'user' WHERE id = ? AND role = 'moderator'`, userID)
			if err != nil {
				http.Error(w, "Error demoting user", http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "Invalid action", http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/users", http.StatusSeeOther)
		return
	}

	// Получаем список пользователей
	rows, err := db.Query("SELECT id, username, email, role FROM users ORDER BY role DESC")
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Role); err != nil {
			http.Error(w, "Error scanning users", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	// Загружаем шаблон
	tmpl, err := template.ParseFS(templates.Files, "users.html")
	if err != nil {
		http.Error(w, "Template parsing error", http.StatusInternalServerError)
		log.Println("Ошибка загрузки шаблона:", err)
		return
	}

	// Отправляем данные в шаблон
	err = tmpl.Execute(w, users)
	if err != nil {
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
		log.Println("Ошибка рендеринга:", err)
	}
}
