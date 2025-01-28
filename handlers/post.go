package handlers

import (
	"forum/database"
	"github.com/gofrs/uuid"
	"html/template"
	"net/http"
	"strconv"
)

// PostHandler обрабатывает отображение конкретного поста
func PostHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем параметр ID поста из URL
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		// Если сессии нет, создаем новую
		sessionID = &http.Cookie{
			Name:  "session_id",
			Value: uuid.Must(uuid.NewV4()).String(), // Используем gofrs/uuid для генерации UUID
		}
		http.SetCookie(w, sessionID)
	}

	username, loggedIn := store[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Получаем ID пользователя по имени
	UserID, ok := id[sessionID.Value]
	if !ok {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Получаем ID поста из параметров URL
	postIDStr := r.URL.Query().Get("id")
	if postIDStr == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// Открываем соединение с базой данных
	db, err := database.InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Получаем пост из базы данных
	post, err := database.GetPostByID(db, postID)
	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Проверяем, является ли текущий пользователь автором поста
	creator := false
	if post.AuthorID == UserID {
		creator = true
	}

	// Обрабатываем POST запрос для удаления поста
	if r.Method == http.MethodPost && creator {
		// Удаляем пост
		err := database.DeletePost(db, postID)
		if err != nil {
			http.Error(w, "Error deleting post", http.StatusInternalServerError)
			return
		}

		// Перенаправляем на страницу с постами
		http.Redirect(w, r, "/my-posts", http.StatusSeeOther)
		return
	}

	// Загружаем шаблон и передаем данные
	tmpl, err := template.ParseFiles("templates/post.html")
	if err != nil {
		http.Error(w, "Template parsing error", http.StatusInternalServerError)
		return
	}

	// Данные для шаблона
	datas := struct {
		LoggedIn bool
		Creator  bool
		Username string
		Post     database.Post
	}{
		LoggedIn: loggedIn,
		Creator:  creator,
		Username: username,
		Post:     post,
	}

	// Прямо вызываем шаблон без дополнительных вызовов WriteHeader
	err = tmpl.Execute(w, datas)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

func UserPostHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем сессию
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		// Если сессии нет
		sessionID = &http.Cookie{
			Name:  "session_id",
			Value: uuid.Must(uuid.NewV4()).String(),
		}
		http.SetCookie(w, sessionID)
		http.Error(w, "User is not logged in", http.StatusUnauthorized)
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
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Открываем соединение с базой данных
	db, err := database.InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// Фильтрация по категориям
	categoryID := r.URL.Query().Get("category_id")

	var posts []database.Post
	if categoryID != "" {
		posts, err = database.GetPostsByUserIDAndCategory(db, UserID, categoryID)
	} else {
		posts, err = database.GetPostsByUserID(db, UserID)
	}
	if err != nil {
		http.Error(w, "Error retrieving user's posts", http.StatusInternalServerError)
		return
	}

	categories, err := database.GetCategories(db)
	if err != nil {
		http.Error(w, "Error fetching categories", http.StatusInternalServerError)
		return
	}

	// Загружаем шаблон для отображения постов
	tmpl, err := template.ParseFiles("templates/my_posts.html")
	if err != nil {
		http.Error(w, "Template parsing error", http.StatusInternalServerError)
		return
	}

	// Данные для шаблона
	data := struct {
		LoggedIn   bool
		ID         int
		Username   string
		Posts      []database.Post
		Categories []database.Category
	}{
		LoggedIn:   loggedIn,
		ID:         UserID,
		Username:   username,
		Posts:      posts,
		Categories: categories,
	}

	// Рендерим шаблон
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
