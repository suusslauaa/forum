package database

import (
	"database/sql"
	"fmt"
	"log"
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
func CreatePost(db *sql.DB, title, content string, authorID, categoryID int, createdAt, imagePath string) error {
	query := `INSERT INTO posts (title, content, author_id, category_id, created_at, image_path) 
			  VALUES (?, ?, ?, ?, ?, ?);`

	// Используем подготовленный запрос для безопасности
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()
	// Выполняем запрос
	log.Println("Executing query:", query)
	log.Println("Params:", title, content, authorID, categoryID, createdAt, imagePath)

	var imgPath interface{}
	if imagePath == "" {
		imgPath = nil
	} else {
		imgPath = imagePath
	}

	_, err = stmt.Exec(title, content, authorID, categoryID, createdAt, imgPath)
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

// ToggleLike переключает состояние лайка для поста
func ToggleLike(db *sql.DB, postID, userID int) error {
	// Проверяем, поставил ли пользователь лайк
	var reactionType string
	err := db.QueryRow("SELECT reaction_type FROM post_reactions WHERE post_id = ? AND user_id = ?", postID, userID).Scan(&reactionType)
	if err == sql.ErrNoRows {
		// Если лайк не поставлен, ставим лайк
		_, err := db.Exec("INSERT INTO post_reactions (post_id, user_id, reaction_type) VALUES (?, ?, 'like')", postID, userID)
		if err != nil {
			return err
		}
		// Увеличиваем счетчик лайков
		_, err = db.Exec("UPDATE posts SET liked = liked + 1 WHERE id = ?", postID)
		return err
	} else if err != nil {
		return err
	}

	if reactionType == "like" {
		// Если пользователь уже поставил лайк, удаляем лайк
		_, err := db.Exec("DELETE FROM post_reactions WHERE post_id = ? AND user_id = ?", postID, userID)
		if err != nil {
			return err
		}
		// Уменьшаем счетчик лайков
		_, err = db.Exec("UPDATE posts SET liked = liked - 1 WHERE id = ?", postID)
		return err
	} else {
		// Если пользователь поставил дизлайк, меняем на лайк
		_, err := db.Exec("UPDATE post_reactions SET reaction_type = 'like' WHERE post_id = ? AND user_id = ?", postID, userID)
		if err != nil {
			return err
		}
		// Обновляем счетчики
		_, err = db.Exec("UPDATE posts SET liked = liked + 1, disliked = disliked - 1 WHERE id = ?", postID)
		return err
	}
}

// ToggleDislike переключает состояние дизлайка для поста
func ToggleDislike(db *sql.DB, postID, userID int) error {
	// Проверяем, поставил ли пользователь дизлайк
	var reactionType string
	err := db.QueryRow("SELECT reaction_type FROM post_reactions WHERE post_id = ? AND user_id = ?", postID, userID).Scan(&reactionType)
	if err == sql.ErrNoRows {
		// Если лайк не поставлен, ставим лайк
		_, err := db.Exec("INSERT INTO post_reactions (post_id, user_id, reaction_type) VALUES (?, ?, 'dislike')", postID, userID)
		if err != nil {
			return err
		}
		// Увеличиваем счетчик лайков
		_, err = db.Exec("UPDATE posts SET disliked = disliked + 1 WHERE id = ?", postID)
		return err
	} else if err != nil {
		return err
	}

	if reactionType == "dislike" {
		// Если пользователь уже поставил лайк, удаляем лайк
		_, err := db.Exec("DELETE FROM post_reactions WHERE post_id = ? AND user_id = ?", postID, userID)
		if err != nil {
			return err
		}
		// Уменьшаем счетчик лайков
		_, err = db.Exec("UPDATE posts SET disliked = disliked - 1 WHERE id = ?", postID)
		return err
	} else {
		// Если пользователь поставил дизлайк, меняем на лайк
		_, err := db.Exec("UPDATE post_reactions SET reaction_type = 'dislike' WHERE post_id = ? AND user_id = ?", postID, userID)
		if err != nil {
			return err
		}
		// Обновляем счетчики
		_, err = db.Exec("UPDATE posts SET disliked = disliked + 1, liked = liked - 1 WHERE id = ?", postID)
		return err
	}
}

func DeletePost(db *sql.DB, postID int) error {
	query := `DELETE FROM posts WHERE id = ?`
	_, err := db.Exec(query, postID)
	return err
}

func AddComment(db *sql.DB, postID, userID int, content string) error {
	// Вставляем новый комментарий в таблицу
	_, err := db.Exec(`
		INSERT INTO comments (post_id, user_id, content)
		VALUES (?, ?, ?)`,
		postID, userID, content,
	)
	return err
}
