package handlers

import (
	"forum/database"
	"forum/templates"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func EditPostHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := GetSessionID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	_, loggedIn := store[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
	}
	loggedInUserID := id[sessionID.Value]

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

	// Если метод GET, отображаем форму для создания поста
	if r.Method == http.MethodGet {
		// Получаем список категорий из базы данных
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

		categories, err := database.GetCategories(db)
		if err != nil {
			ErrorHandler(w, "Error fetching categories", http.StatusInternalServerError)
			return
		}

		// Передаем категории в шаблон
		tmpl, err := template.ParseFS(templates.Files, "edit_post.html")
		if err != nil {
			ErrorHandler(w, "Template parsing error", http.StatusInternalServerError)
			return
		}
		data := struct {
			Post       database.Post
			UserID     int
			Categories []database.Category
			Check      string
		}{
			UserID:     loggedInUserID, // Получите ID пользователя из сессии
			Categories: categories,
			Post:       post,
			Check:      checker,
		}
		if checker != "" {
			checker = ""
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			ErrorHandler(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method == http.MethodPost {
		title := r.FormValue("title")
		content := r.FormValue("content")
		userID := r.FormValue("user_id")
		categoryIDstr := r.FormValue("category")
		var categoryID *int
		if categoryIDstr == "" {
			categoryID = nil
		} else {
			ID, err := strconv.Atoi(categoryIDstr)
			if err != nil {
				ErrorHandler(w, "Invalid category ID", http.StatusBadRequest)
				return
			}
			categoryID = &ID
		}

		if title == "" || content == "" || userID == "" {
			checker = "All fields are required"
			http.Redirect(w, r, "/edit-post?$="+strconv.Itoa(postID), http.StatusSeeOther)
			return
		}

		// Открываем соединение с базой данных
		db, err := database.InitDB()
		if err != nil {
			ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		// Парсим форму, ограничив размер файла (например, 10MB)
		err = r.ParseMultipartForm(10 << 20) // 10 MB
		if err != nil {
			http.Error(w, "File too large", http.StatusBadRequest)
			return
		}

		// Получаем файл из формы
		file, handler, err := r.FormFile("image")
		var savePath string
		// Проверяем, был ли файл загружен
		if err != nil && err != http.ErrMissingFile {
			http.Error(w, "Error retrieving file", http.StatusBadRequest)
			return
		}
		if file != nil {
			defer file.Close()

			// Опционально: проверка расширения
			allowedExtensions := map[string]bool{
				".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
			}
			ext := filepath.Ext(handler.Filename)
			if !allowedExtensions[ext] {
				http.Error(w, "Invalid file type", http.StatusBadRequest)
				return
			}

			// Создаем файл на сервере
			savePath = filepath.Join("uploads", handler.Filename)
			outFile, err := os.Create(savePath)
			if err != nil {
				http.Error(w, "Unable to save file", http.StatusInternalServerError)
				return
			}
			defer outFile.Close()
			// Копируем содержимое файла
			_, err = io.Copy(outFile, file)
			if err != nil {
				http.Error(w, "Failed to save file", http.StatusInternalServerError)
				return
			}
		} else {
			// Если файл не был загружен, оставляем savePath пустым
			savePath = ""
		}

		// Вставляем пост в базу данных
		createdAt := time.Now().Format("2006-01-02 15:04:05")
		err = database.EditPost(db, title, content, categoryID, createdAt, postID, savePath)
		if err != nil {
			ErrorHandler(w, "Error saving post to database", http.StatusInternalServerError)
			return
		}

		// Перенаправляем на главную страницу или список постов
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
