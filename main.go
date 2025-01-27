package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // Импорт SQLite драйвера
	"golang.org/x/crypto/bcrypt"
)

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		return nil, err
	}

	createTable := `
    CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        username TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL
    );
    `
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	// Создание таблицы категорий
	createCategoryTableSQL := `CREATE TABLE IF NOT EXISTS categories (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"name" TEXT NOT NULL UNIQUE
	);`
	_, err = db.Exec(createCategoryTableSQL)
	if err != nil {
		return nil, err
	}

	// Создание новой таблицы постов с category_id
	createPostTableSQL := `CREATE TABLE IF NOT EXISTS posts_new (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"title" TEXT,
		"content" TEXT,
		"category_id" INTEGER,
		FOREIGN KEY (category_id) REFERENCES categories(id)
	);`
	_, err = db.Exec(createPostTableSQL)
	if err != nil {
		return nil, err
	}

	// Перенос данных из старой таблицы в новую (если таблица posts уже существует)
	_, err = db.Exec(`
		INSERT INTO posts_new (id, title, content, category_id)
		SELECT id, title, content, category_id FROM posts
	`)
	if err != nil {
		return nil, err
	}

	// Удаление старой таблицы
	_, err = db.Exec(`DROP TABLE IF EXISTS posts`)
	if err != nil {
		return nil, err
	}

	// Переименование новой таблицы в posts
	_, err = db.Exec(`ALTER TABLE posts_new RENAME TO posts`)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func clearTable(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM posts")
	return err
}

func createCategory(db *sql.DB, name string) error {
	_, err := db.Exec(`INSERT INTO categories (name) VALUES (?)`, name)
	return err
}

func deleteCategory(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM categories WHERE id = ?`, id)
	return err
}

func createPost(db *sql.DB, title string, content string, categoryId int) error {
	// Проверка на существование поста с таким же заголовком и содержанием
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE title = ? AND content = ?)`
	err := db.QueryRow(query, title, content).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil // Если пост уже существует, ничего не делаем
	}

	// Вставка нового поста
	insertPostSQL := `INSERT INTO posts (title, content, category_id) VALUES (?, ?, ?)`
	_, err = db.Exec(insertPostSQL, title, content, categoryId)
	return err
}

func main() {
	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// clearTable(db)

	createCategory(db, "General")
	createCategory(db, "Technology")

	err = createPost(db, "First Post", "This is the content of the first post.", 1)
	if err != nil {
		log.Fatal(err)
	}

	err = createPost(db, "Second Post", "This is the content of the second post.", 2)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/account", profileHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", homeHandler)
	http.ListenAndServe(":8080", nil)
}

type Post struct {
	ID       int
	Title    string
	Content  string
	Category string
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	db, err := initDB()
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

	posts, err := getPosts(db, categoryID)
	if err != nil {
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}

	categories, err := getCategories(db)
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

func getPosts(db *sql.DB, categoryID int) ([]Post, error) {
	var rows *sql.Rows
	var err error

	if categoryID > 0 {
		rows, err = db.Query(`
			SELECT p.id, p.title, p.content, c.name 
			FROM posts p
			LEFT JOIN categories c ON p.category_id = c.id
			WHERE p.category_id = ?`, categoryID)
	} else {
		rows, err = db.Query(`
			SELECT p.id, p.title, p.content, c.name 
			FROM posts p
			LEFT JOIN categories c ON p.category_id = c.id`)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var categoryName string
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &categoryName)
		if err != nil {
			return nil, err
		}
		post.Category = categoryName
		posts = append(posts, post)
	}

	return posts, nil
}

func getCategories(db *sql.DB) ([]Category, error) {
	rows, err := db.Query(`SELECT id, name FROM categories`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

type Category struct {
	ID   int
	Name string
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
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

		db, err := initDB()
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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		db, err := initDB()
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

func profileHandler(w http.ResponseWriter, r *http.Request) {
	// В реальном приложении здесь бы проверялась сессия пользователя
	// Здесь просто показываем страницу профиля

	http.ServeFile(w, r, "./templates/account.html")
}
