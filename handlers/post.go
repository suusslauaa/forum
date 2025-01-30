package handlers

import (
	"database/sql"
	"fmt"
	"forum/database"
	"forum/templates"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

// PostHandler обрабатывает отображение конкретного поста
func PostHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка сессии
	sessionID, err := GetSessionID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	username, loggedIn := store[sessionID.Value]

	// Получаем ID пользователя по имени
	UserID, _ := id[sessionID.Value]
	var role string
	if loggedIn {
		UserID = id[sessionID.Value]

		// Получаем роль пользователя
		db, err := database.InitDB()
		if err != nil {
			ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		role, err = GetUserRole(db, UserID)
		if err != nil {
			ErrorHandler(w, "Error fetching user role", http.StatusInternalServerError)
			return
		}
	}
	Moders := false
	if role == "moderator" || role == "admin" {
		Moders = true
	}
	admin := false
	if role == "admin" {
		admin = true
	}
	fmt.Println(role, Moders)
	// Получаем ID поста из параметров URL
	postIDStr := r.URL.Query().Get("id")

	if postIDStr == "" {
		ErrorHandler(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		ErrorHandler(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Открываем соединение с базой данных
	db, err := database.InitDB()
	if err != nil {
		ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Получаем пост из базы данных
	post, err := database.GetPostByID(db, postID)
	if err != nil {
		ErrorHandler(w, "Post not found", http.StatusNotFound)
		return
	}

	// Получаем комментарии для поста
	comments := post.Comments

	// Проверяем, является ли текущий пользователь автором поста
	creator := post.AuthorID == UserID

	// Обрабатываем POST-запрос для добавления комментария
	if r.Method == http.MethodPost && loggedIn {

		if err := handlePostActions(w, r, db, postID, UserID, Moders); err != nil {
			ErrorHandler(w, "Error handling post action", http.StatusInternalServerError)
			return
		}

		// Перенаправляем на страницу поста с добавленным комментарием
		http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
		return
	}

	// Загружаем шаблон и передаем данные
	tmpl, err := template.ParseFS(templates.Files, "post.html")
	if err != nil {
		ErrorHandler(w, "Template parsing error", http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"LoggedIn":      loggedIn,
		"Creator":       creator,
		"Username":      username,
		"Post":          post,
		"Comments":      comments,
		"SessionUserID": UserID,
		"Moder":         Moders,
		"Admin":         admin,
		"UserRole":      role,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		ErrorHandler(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

// Обрабатываем лайки, дизлайки и удаление
func handlePostActions(w http.ResponseWriter, r *http.Request, db *sql.DB, postID int, userID int, moders bool) error {
	action := r.URL.Query().Get("action")
	switch action {
	case "like":
		if err := database.ToggleLike(db, postID, userID); err != nil {
			ErrorHandler(w, "Error processing like", http.StatusInternalServerError)
			return err
		}
	case "dislike":
		if err := database.ToggleDislike(db, postID, userID); err != nil {
			ErrorHandler(w, "Error processing dislike", http.StatusInternalServerError)
			return err
		}
	case "delete":
		if err := database.DeletePost(db, postID); err != nil {
			ErrorHandler(w, "Error deleting post", http.StatusInternalServerError)
			return err
		}
		database.DeletePostReport(db, postID)
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	case "comment":
		commentText := r.FormValue("content")
		if commentText == "" {
			ErrorHandler(w, "Comment text is required", http.StatusBadRequest)
			return nil
		}
		if err := database.AddComment(db, postID, userID, commentText); err != nil {
			ErrorHandler(w, "Error adding comment", http.StatusInternalServerError)
			return err
		}
	case "report":
		if !moders {
			ErrorHandler(w, "Error reporting post", http.StatusInternalServerError)
			return nil
		}
		err := database.ReportPost(db, postID, userID)
		if err != nil {
			ErrorHandler(w, "Database error", http.StatusInternalServerError)
			log.Println("Error updating report:", err)
			return nil
		}
	default:
		ErrorHandler(w, "Invalid action", http.StatusBadRequest)
		return nil
	}

	return nil
}

func CommentHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка сессии
	sessionID, err := GetSessionID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	_, loggedIn := store[sessionID.Value]

	// Получаем ID пользователя по имени
	UserID, _ := id[sessionID.Value]

	// Получаем ID поста из параметров URL
	commentIDstr := r.URL.Query().Get("id")
	if commentIDstr == "" {
		ErrorHandler(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(commentIDstr)
	if err != nil {
		ErrorHandler(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Открываем соединение с базой данных
	db, err := database.InitDB()
	if err != nil {
		ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	postID, _ := database.GetPostIDByCommentID(db, commentID)
	if r.Method == http.MethodPost && loggedIn {

		if err := handleCommentActions(w, r, db, commentID, UserID); err != nil {
			ErrorHandler(w, "Error handling post action", http.StatusInternalServerError)
			return
		}

		// Перенаправляем на страницу поста с добавленным комментарием
		http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
		return
	}
}
