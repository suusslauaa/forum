package database

import (
	"database/sql"
	"fmt"
)

// ClearTable удаляет все записи из таблицы posts
func ClearTable(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM posts")
	return err
}

// CreateCategory создает новую категорию
func CreateCategory(db *sql.DB, name string) error {
	_, err := db.Exec(`INSERT INTO categories (name) VALUES (?)`, name)
	return err
}

func EditPost(db *sql.DB, title, content string, categoryID int, createdAt string, id int) error {
	query := `
		UPDATE posts 
		SET title = ?, content = ?, category_id = ?, created_at = ?
		WHERE id = ?;
	`

	// Используем подготовленный запрос для безопасности
	stmt, err := db.Prepare(query)
	if err != nil {
		fmt.Println("Ошибка при подготовке запроса:", err)
		return err
	}
	defer stmt.Close()

	// Выполняем запрос с передачей ID записи
	_, err = stmt.Exec(title, content, categoryID, createdAt, id)
	if err != nil {
		fmt.Println("Ошибка выполнения запроса:", err)
	}
	return err
}

// DeleteCategory удаляет категорию по id
func DeleteCategory(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM categories WHERE id = ?`, id)
	return err
}

// CreatePost создает новый пост, привязанный к категории
func CreatePost(db *sql.DB, title, content string, authorID, categoryID int, createdAt string, liked int) error {
	query := `INSERT INTO posts (title, content, author_id, category_id, created_at,liked) 
			  VALUES (?, ?, ?, ?, ?, ?);`

	// Используем подготовленный запрос для безопасности
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	// Выполняем запрос
	_, err = stmt.Exec(title, content, authorID, categoryID, createdAt, liked)
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

func DeletePost(db *sql.DB, postID int) error {
	query := `DELETE FROM posts WHERE id = ?`
	_, err := db.Exec(query, postID)
	return err
}
