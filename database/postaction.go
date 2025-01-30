package database

import (
	"database/sql"
	"fmt"
	"strconv"
)

func EditPost(db *sql.DB, title, content string, categoryID *int, createdAt string, id int, savePath string) error {
	query := `
		UPDATE posts 
		SET title = ?, content = ?, category_id = ?, created_at = ?, image_path = ?
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
	_, err = stmt.Exec(title, content, categoryID, createdAt, savePath, id)
	if err != nil {
		fmt.Println("Ошибка выполнения запроса:", err)
	}
	return err
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
		query := `INSERT INTO activities (user_id, activity_type, post_id, created_at) 
		VALUES ($1, 'You liked', $2, CURRENT_TIMESTAMP)`
		_, err = db.Exec(query, userID, postID)

		post, _ := GetPostByID(db, postID)
		if userID != post.AuthorID {
			query = `
			INSERT INTO notifications (user_id, notification_type, post_id, created_at) 
			VALUES ($1, 'You received a like', $2, CURRENT_TIMESTAMP)
			`
			post, _ := GetPostByID(db, postID)
			_, err = db.Exec(query, post.AuthorID, postID)
		}

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
		query := `INSERT INTO activities (user_id, activity_type, post_id, created_at) 
		VALUES ($1, 'You liked', $2, CURRENT_TIMESTAMP)`
		_, err = db.Exec(query, userID, postID)
		post, _ := GetPostByID(db, postID)
		if userID != post.AuthorID {
			query = `INSERT INTO notifications (user_id, notification_type, post_id, created_at) 
		VALUES ($1, 'You received a like', $2, CURRENT_TIMESTAMP)`
			post, _ := GetPostByID(db, postID)
			_, err = db.Exec(query, post.AuthorID, postID)
		}
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
		query := `INSERT INTO activities (user_id, activity_type, post_id, created_at) 
		VALUES ($1, 'You disliked', $2, CURRENT_TIMESTAMP)`
		_, err = db.Exec(query, userID, postID)
		post, _ := GetPostByID(db, postID)
		if userID != post.AuthorID {
			query = `INSERT INTO notifications (user_id, notification_type, post_id, created_at) 
		VALUES ($1, 'You received a dislike', $2, CURRENT_TIMESTAMP)`
			post, _ := GetPostByID(db, postID)
			_, err = db.Exec(query, post.AuthorID, postID)
		}
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
		query := `INSERT INTO activities (user_id, activity_type, post_id, created_at) 
		VALUES ($1, 'You disliked', $2, CURRENT_TIMESTAMP)`
		_, err = db.Exec(query, userID, postID)
		post, _ := GetPostByID(db, postID)
		if userID != post.AuthorID {
			query = `INSERT INTO notifications (user_id, notification_type, post_id, created_at) 
		VALUES ($1, 'You received a dislike', $2, CURRENT_TIMESTAMP)`
			post, _ := GetPostByID(db, postID)
			_, err = db.Exec(query, post.AuthorID, postID)
		}

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
	query := `INSERT INTO comments (post_id, user_id, content)
			VALUES (?, ?, ?) RETURNING id`

	// Используем подготовленный запрос для безопасности
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}

	defer stmt.Close()

	var commentID string

	// Выполняем запрос и получаем ID вставленного поста
	err = db.QueryRow(query, postID, userID, content).Scan(&commentID)
	if err != nil {
		return err
	}

	query = `INSERT INTO activities (user_id, activity_type, comment_id, comment_content, post_id, created_at) 
	           VALUES ($1, 'You commented', $2, $3, $4, CURRENT_TIMESTAMP)`
	_, err = db.Exec(query, userID, commentID, content, postID)
	com, _ := strconv.Atoi(commentID)
	comAuth, _ := GetUserIDByCommentID(db, com)
	post, _ := GetPostByID(db, postID)
	if comAuth != post.AuthorID {
		query = `INSERT INTO notifications (user_id, notification_type, post_id, comment_content, created_at) 
	VALUES ($1, 'You received a comment', $2, $3, CURRENT_TIMESTAMP)`

		_, err = db.Exec(query, post.AuthorID, post.ID, content, postID)
	}
	return err
}
