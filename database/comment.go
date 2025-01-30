package database

import "database/sql"

func ToggleDislikeComment(db *sql.DB, commentID, userID int) error {
	var reactionType string
	err := db.QueryRow("SELECT reaction_type FROM comment_reactions WHERE comment_id = ? AND user_id = ?", commentID, userID).Scan(&reactionType)
	if err == sql.ErrNoRows {
		// Если лайк не поставлен, ставим лайк
		_, err := db.Exec("INSERT INTO comment_reactions (comment_id, user_id, reaction_type) VALUES (?, ?, 'dislike')", commentID, userID)
		if err != nil {
			return err
		}
		// Увеличиваем счетчик лайков
		_, err = db.Exec("UPDATE comments SET disliked = disliked + 1 WHERE id = ?", commentID)
		return err
	} else if err != nil {
		return err
	}

	if reactionType == "dislike" {
		// Если пользователь уже поставил лайк, удаляем лайк
		_, err := db.Exec("DELETE FROM comment_reactions WHERE comment_id = ? AND user_id = ?", commentID, userID)
		if err != nil {
			return err
		}
		// Уменьшаем счетчик лайков
		_, err = db.Exec("UPDATE comments SET disliked = disliked - 1 WHERE id = ?", commentID)
		return err
	} else {
		// Если пользователь поставил дизлайк, меняем на лайк
		_, err := db.Exec("UPDATE comment_reactions SET reaction_type = 'dislike' WHERE comment_id = ? AND user_id = ?", commentID, userID)
		if err != nil {
			return err
		}
		// Обновляем счетчики
		_, err = db.Exec("UPDATE comments SET disliked = disliked + 1, liked = liked - 1 WHERE id = ?", commentID)
		return err
	}
}

func ToggleLikeComment(db *sql.DB, commentID, userID int) error {
	var reactionType string
	err := db.QueryRow("SELECT reaction_type FROM comment_reactions WHERE comment_id = ? AND user_id = ?", commentID, userID).Scan(&reactionType)
	if err == sql.ErrNoRows {
		// Если лайк не поставлен, ставим лайк
		_, err := db.Exec("INSERT INTO comment_reactions (comment_id, user_id, reaction_type) VALUES (?, ?, 'like')", commentID, userID)
		if err != nil {
			return err
		}
		// Увеличиваем счетчик лайков
		_, err = db.Exec("UPDATE comments SET liked = liked + 1 WHERE id = ?", commentID)
		return err
	} else if err != nil {
		return err
	}

	if reactionType == "like" {
		// Если пользователь уже поставил лайк, удаляем лайк
		_, err := db.Exec("DELETE FROM comment_reactions WHERE comment_id = ? AND user_id = ?", commentID, userID)
		if err != nil {
			return err
		}
		// Уменьшаем счетчик лайков
		_, err = db.Exec("UPDATE comments SET liked = liked - 1 WHERE id = ?", commentID)
		return err
	} else {
		// Если пользователь поставил дизлайк, меняем на лайк
		_, err := db.Exec("UPDATE comment_reactions SET reaction_type = 'like' WHERE comment_id = ? AND user_id = ?", commentID, userID)
		if err != nil {
			return err
		}
		// Обновляем счетчики
		_, err = db.Exec("UPDATE comments SET liked = liked + 1, disliked = disliked - 1 WHERE id = ?", commentID)
		return err
	}
}

func DeleteComment(db *sql.DB, commentID int) error {
	query := `DELETE FROM comments WHERE id = ?`
	_, err := db.Exec(query, commentID)
	return err
}

func UpdateComment(db *sql.DB, commentID int, userID int, newContent string) error {
	query := `UPDATE comments SET content = ? WHERE id = ? AND user_id = ?`
	_, err := db.Exec(query, newContent, commentID, userID)
	return err
}
