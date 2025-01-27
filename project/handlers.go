package project

import (
	"log"
	"net/http"
	"strconv"
	"text/template"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	db, err := InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	categoryID := 0
	if r.URL.Query().Get("category_id") != "" {
		categoryID, err = strconv.Atoi(r.URL.Query().Get("category_id"))
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}
	}

	posts, err := GetPosts(db, categoryID)
	if err != nil {
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}

	categories, err := GetCategories(db)
	if err != nil {
		http.Error(w, "Error fetching categories", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("./templates/home.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Posts      []Post
		Categories []Category
	}{
		Posts:      posts,
		Categories: categories,
	}

	tmpl.Execute(w, data)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Хэшируем пароль
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		db, err := InitDB()
		if err != nil {
			log.Fatal(err)
		}

		defer db.Close()

		// Генерируем UUID для пользователя
		userID := uuid.New().String()

		// Вставляем данные пользователя в базу данных
		query := `INSERT INTO users (id, username, email, password) VALUES (?, ?, ?, ?)`
		_, err = db.Exec(query, userID, username, email, hashedPassword)
		if err != nil {
			http.Error(w, "Error inserting user into database", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.ServeFile(w, r, "./templates/register.html")
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		db, err := InitDB()
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// Проверяем пользователя по email
		query := `SELECT id, password FROM users WHERE email = ?`
		row := db.QueryRow(query, email)

		var storedPassword string
		var userID string
		if err := row.Scan(&userID, &storedPassword); err != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Сравниваем пароли
		err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
		if err != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Успешный вход — создаем сессию или просто показываем страницу профиля
		// В реальном приложении можно использовать cookie или сессии для хранения состояния

		http.Redirect(w, r, "/account", http.StatusSeeOther)
		return
	}

	http.ServeFile(w, r, "./templates/login.html")
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	// В реальном приложении здесь бы проверялась сессия пользователя
	// Здесь просто показываем страницу профиля

	db, err := InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	categoryID := 0
	if r.URL.Query().Get("category_id") != "" {
		categoryID, err = strconv.Atoi(r.URL.Query().Get("category_id"))
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}
	}

	posts, err := GetPosts(db, categoryID)
	if err != nil {
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}

	categories, err := GetCategories(db)
	if err != nil {
		http.Error(w, "Error fetching categories", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("./templates/account.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Username   string
		Posts      []Post
		Categories []Category
	}{
		Username:   "kk",
		Posts:      posts,
		Categories: categories,
	}

	tmpl.Execute(w, data)
}
