package static

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var DB *sql.DB

// InitDB инициализирует базу данных и создает таблицы
// InitDB инициализирует базу данных и создает таблицы
func InitDB() (*sql.DB, error) {
	// Открытие базы данных
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		return nil, err
	}

	// Сохраняем глобальную переменную DB
	DB = db

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

	// Вызов функции для создания других таблиц
	createTables(db)

	return db, nil
}

// createTables создает необходимые таблицы в базе данных
func createTables(db *sql.DB) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT UNIQUE NOT NULL,
		user_id INTEGER NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id)
		);`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Fatalf("Ошибка создания таблиц: %v", err)
		}
	}
}

func CheckEmailExists(email string) (bool, error) {
	// Убедитесь, что соединение с базой данных открыто
	db, err := InitDB()
	if err != nil {
		return false, err
	}
	defer db.Close()

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	if err != nil {
		return false, err
	}

	// Если count > 0, то email уже существует

	return count > 0, nil
}
