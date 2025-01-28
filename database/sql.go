package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var DB *sql.DB

// InitDB инициализирует базу данных и создает таблицы
func InitDB() (*sql.DB, error) {
	// Открытие базы данных
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		return nil, err
	}

	// Сохраняем глобальное соединение
	DB = db

	// Создаем таблицы, если их еще нет
	err = createTables(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// createTables создает необходимые таблицы
func createTables(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		);`,
		`CREATE TABLE IF NOT EXISTS posts (
			id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			category_id INTEGER,
			author_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			liked INTEGER DEFAULT 0, -- Начальное значение лайков 0
			disliked INTEGER DEFAULT 0, -- Начальное значение лайков 0
			FOREIGN KEY (category_id) REFERENCES categories(id),
			FOREIGN KEY (author_id) REFERENCES users(id)
		);`,
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
		`CREATE TABLE IF NOT EXISTS post_reactions (
			post_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			reaction_type TEXT CHECK(reaction_type IN ('like', 'dislike')) NOT NULL, -- Сохраняем тип реакции как текст
			PRIMARY KEY (post_id, user_id),
			FOREIGN KEY (post_id) REFERENCES posts(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS comments (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL,            -- Ссылка на пост
    user_id INTEGER NOT NULL,            -- Ссылка на пользователя (автора комментария)
    content TEXT NOT NULL,               -- Содержание комментария
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,  -- Время создания комментария
    FOREIGN KEY (post_id) REFERENCES posts(id),      -- Внешний ключ на таблицу постов
    FOREIGN KEY (user_id) REFERENCES users(id)       -- Внешний ключ на таблицу пользователей
);`,
	}

	// Выполнение всех запросов на создание таблиц
	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Printf("Ошибка выполнения запроса: %s\n%v", query, err)
			return err
		}
	}

	return nil
}

// CheckEmailExists проверяет, существует ли email в базе
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
