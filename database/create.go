package database

import (
	"database/sql"
	"fmt"
)

// ClearTable удаляет все записи из таблицы posts
func ClearTable(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM posts_new")
	return err
}

// CreateCategory создает новую категорию
func CreateCategory(db *sql.DB, name string) error {
	_, err := db.Exec(`INSERT INTO categories (name) VALUES (?)`, name)
	return err
}

// DeleteCategory удаляет категорию по id
func DeleteCategory(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM categories WHERE id = ?`, id)
	return err
}

// CreatePost создает новый пост, привязанный к категории
func CreatePost(db *sql.DB, title, content string, authorID, categoryID int, createdAt string) error {
	query := `INSERT INTO posts (title, content, author_id, category_id, created_at) 
			  VALUES (?, ?, ?, ?, ?);`

	// Используем подготовленный запрос для безопасности
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Выполняем запрос
	_, err = stmt.Exec(title, content, authorID, categoryID, createdAt)
	return err
}

func CreateUser(db *sql.DB, email, username, password string) error {
	// SQL-запрос для вставки пользователя в таблицу
	query := `INSERT INTO users (email, username, password) VALUES (?, ?, ?);`

	// Подготовка запроса
	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	// Выполнение запроса
	_, err = stmt.Exec(email, username, password)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	return nil
}
