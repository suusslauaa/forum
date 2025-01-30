package handlers

import (
	"database/sql"
	"forum/database"
	"net/http"
)

func handleCommentActions(w http.ResponseWriter, r *http.Request, db *sql.DB, commentID int, userID int) error {
	action := r.URL.Query().Get("action")

	switch action {
	case "like":
		if err := database.ToggleLikeComment(db, commentID, userID); err != nil {
			ErrorHandler(w, "Error processing like for comment", http.StatusInternalServerError)
			return err
		}
	case "dislike":
		if err := database.ToggleDislikeComment(db, commentID, userID); err != nil {
			ErrorHandler(w, "Error processing dislike for comment", http.StatusInternalServerError)
			return err
		}
	case "delete":
		if err := database.DeleteComment(db, commentID); err != nil {
			ErrorHandler(w, "Error deleting comment", http.StatusInternalServerError)
			return err
		}
	case "update":
		newContent := r.FormValue("content")
		if newContent == "" {
			ErrorHandler(w, "Comment content is required", http.StatusBadRequest)
			return nil
		}
		if err := database.UpdateComment(db, commentID, userID, newContent); err != nil {
			ErrorHandler(w, "Error updating comment", http.StatusInternalServerError)
			return err
		}
	default:
		ErrorHandler(w, "Invalid comment action", http.StatusBadRequest)
		return nil
	}

	return nil
}
