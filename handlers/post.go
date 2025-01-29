package handlers

import (
	"database/sql"
	"forum/database"
	"github.com/gofrs/uuid"
	"html/template"
	"net/http"
	"strconv"
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
		if err := handlePostActions(w, r, db, postID, creator, UserID); err != nil {
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
		LoggedIn bool
		Creator  bool
		Username string
		Post     database.Post
		Comments []database.Comment
	}{
		LoggedIn: loggedIn,
		Creator:  creator,
		Username: username,
		Post:     post,
		Comments: comments,
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
func handlePostActions(w http.ResponseWriter, r *http.Request, db *sql.DB, postID int, creator bool, userID int) error {
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
		if !creator {
			ErrorHandler(w, "You are not the creator of this post", http.StatusForbidden)
			return nil
		}
		if err := database.DeletePost(db, postID); err != nil {
			ErrorHandler(w, "Error deleting post", http.StatusInternalServerError)
			return err
		}
		http.Redirect(w, r, "/my-posts", http.StatusSeeOther)
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
