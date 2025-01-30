package handlers

import (
	"database/sql"
	"forum/database"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/gofrs/uuid"
)

// PostHandler обрабатывает отображение конкретного поста
func PostHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка сессии
	sessionID, err := getSessionID(w, r)
	if err != nil {
		ErrorHandler(w, "Session error", http.StatusInternalServerError)
		return
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
	tmpl, err := template.ParseFiles("templates/post.html")
	if err != nil {
		ErrorHandler(w, "Template parsing error", http.StatusInternalServerError)
		return
	}

	// Передаем данные в шаблон
	datas := struct {
		LoggedIn      bool
		Creator       bool
		Username      string
		Post          database.Post
		Comments      []database.Comment
		SessionUserID int
		Moder         bool
		UserRole      string
	}{
		LoggedIn:      loggedIn,
		Creator:       creator,
		Username:      username,
		Post:          post,
		Comments:      comments,
		SessionUserID: UserID,
		Moder:         Moders,
		UserRole:      role,
	}

	err = tmpl.Execute(w, datas)
	if err != nil {
		ErrorHandler(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

// Функция для получения sessionID из cookies
func getSessionID(w http.ResponseWriter, r *http.Request) (*http.Cookie, error) {
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		sessionID = &http.Cookie{
			Name:  "session_id",
			Value: uuid.Must(uuid.NewV4()).String(),
		}
		http.SetCookie(w, sessionID)
	}
	return sessionID, err
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

func UserPostHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем сессию
	sessionID, err := getSessionID(w, r)
	if err != nil {
		ErrorHandler(w, "Session error", http.StatusInternalServerError)
		return
	}

	// Проверяем, авторизован ли пользователь
	username, loggedIn := store[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Получаем ID пользователя по имени
	UserID, ok := id[sessionID.Value]
	if !ok {
		ErrorHandler(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Открываем соединение с базой данных
	db, err := database.InitDB()
	if err != nil {
		ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Получаем посты пользователя
	posts, err := database.GetPostsByUserID(db, UserID)
	if err != nil {
		ErrorHandler(w, "Error retrieving user's posts", http.StatusInternalServerError)
		return
	}

	// Загружаем шаблон для отображения постов
	tmpl, err := template.ParseFiles("templates/my_posts.html")
	if err != nil {
		ErrorHandler(w, "Template parsing error", http.StatusInternalServerError)
		return
	}

	// Данные для шаблона
	data := struct {
		LoggedIn bool
		ID       int
		Username string
		Posts    []database.Post
	}{
		LoggedIn: loggedIn,
		ID:       UserID,
		Username: username,
		Posts:    posts,
	}

	// Рендерим шаблон
	err = tmpl.Execute(w, data)
	if err != nil {
		ErrorHandler(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func LikePostHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем сессию
	sessionID, err := getSessionID(w, r)
	if err != nil {
		ErrorHandler(w, "Error retrieving session", http.StatusUnauthorized)
		return
	}

	// Проверяем, авторизован ли пользователь
	username, loggedIn := store[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Получаем ID пользователя по имени
	UserID, ok := id[sessionID.Value]
	if !ok {
		ErrorHandler(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Открываем соединение с базой данных
	db, err := database.InitDB()
	if err != nil {
		ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Получаем посты, которые пользователь лайкнул
	posts, err := database.GetLikedPostsByUserID(db, UserID)
	if err != nil {
		ErrorHandler(w, "Error retrieving user's liked posts", http.StatusInternalServerError)
		return
	}

	// Загружаем шаблон для отображения постов
	tmpl, err := template.ParseFiles("templates/my_posts.html")
	if err != nil {
		ErrorHandler(w, "Template parsing error", http.StatusInternalServerError)
		return
	}

	// Данные для шаблона
	data := struct {
		LoggedIn bool
		ID       int
		Username string
		Posts    []database.Post
	}{
		LoggedIn: loggedIn,
		ID:       UserID,
		Username: username,
		Posts:    posts,
	}

	// Рендерим шаблон
	err = tmpl.Execute(w, data)
	if err != nil {
		ErrorHandler(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

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

func handleCommentActions(w http.ResponseWriter, r *http.Request, db *sql.DB, commentID int, userID int) error {
	action := r.URL.Query().Get("action")

	switch action {
	case "like":
		if err := ToggleLikeComment(db, commentID, userID); err != nil {
			ErrorHandler(w, "Error processing like for comment", http.StatusInternalServerError)
			return err
		}
	case "dislike":
		if err := ToggleDislikeComment(db, commentID, userID); err != nil {
			ErrorHandler(w, "Error processing dislike for comment", http.StatusInternalServerError)
			return err
		}
	case "delete":
		if err := DeleteComment(db, commentID); err != nil {
			ErrorHandler(w, "Error deleting comment", http.StatusInternalServerError)
			return err
		}
	case "update":
		newContent := r.FormValue("content")
		if newContent == "" {
			ErrorHandler(w, "Comment content is required", http.StatusBadRequest)
			return nil
		}
		if err := UpdateComment(db, commentID, userID, newContent); err != nil {
			ErrorHandler(w, "Error updating comment", http.StatusInternalServerError)
			return err
		}
	default:
		ErrorHandler(w, "Invalid comment action", http.StatusBadRequest)
		return nil
	}

	return nil
}

func CommentHandler(w http.ResponseWriter, r *http.Request) {
	// Проверка сессии
	sessionID, err := getSessionID(w, r)
	if err != nil {
		ErrorHandler(w, "Session error", http.StatusInternalServerError)
		return
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

func GetUserActivity(w http.ResponseWriter, r *http.Request) {
	// Проверка сессии
	sessionID, err := getSessionID(w, r)
	if err != nil {
		ErrorHandler(w, "Session error", http.StatusInternalServerError)
		return
	}

	// _, loggedIn := store[sessionID.Value]

	// Получаем ID пользователя по имени
	userID, _ := id[sessionID.Value]

	query := `
	SELECT a.activity_type, a.created_at, p.id AS post_id, c.content AS comment_content
	FROM activities a
	LEFT JOIN posts p ON a.post_id = p.id
	LEFT JOIN comments c ON a.comment_id = c.id
	WHERE a.user_id = $1
	ORDER BY a.created_at DESC;
	`
	// Открываем соединение с базой данных
	db, err := database.InitDB()
	if err != nil {
		ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	rows, err := db.Query(query, userID)
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var activities []Activity
	for rows.Next() {
		var activity Activity
		// Сканируем в указатели для корректной работы с NULL
		if err := rows.Scan(&activity.Type, &activity.CreatedAt, &activity.PostID, &activity.CommentContent); err != nil {
			log.Println(err)
			continue
		}
		activities = append(activities, activity)
	}

	tmpl, err := template.ParseFiles("templates/activity_page.html")
	if err != nil {
		log.Println("Error loading template:", err)
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	// Передаем данные в шаблон
	data := map[string]interface{}{
		"Activities": activities,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
		return
	}
}

type Activity struct {
	ID             int
	AuthorID       int
	Type           string
	PostID         *int
	CommentContent *string
	CreatedAt      string
}
