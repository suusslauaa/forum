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

// CreatePostHandler обрабатывает создание нового поста
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	username, loggedIn := store[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	loggedInUserID := id[sessionID.Value]

	// Если метод GET, отображаем форму для создания поста
	if r.Method == http.MethodGet {
		// Получаем список категорий из базы данных
		db, err := database.InitDB()
		if err != nil {
			ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
			return
		}
		defer db.Close()

		categories, err := database.GetCategories(db)
		if err != nil {
			ErrorHandler(w, "Error fetching categories", http.StatusInternalServerError)
			return
		}

		// Передаем категории в шаблон
		tmpl, err := template.ParseFS(templates.Files, "create_post.html")
		if err != nil {
			ErrorHandler(w, "Template parsing error", http.StatusInternalServerError)
			return
		}

		data := struct {
			Username   string
			UserID     int
			Categories []database.Category
			Check      string
		}{
			Username:   username,
			UserID:     loggedInUserID, // Получите ID пользователя из сессии
			Categories: categories,
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
		userID := id[sessionID.Value]

		categoryStr := r.FormValue("category")
		var categoryID *int // Указатель на int (чтобы можно было передавать nil)
		if categoryStr != "" {
			catID, err := strconv.Atoi(categoryStr)
			if err != nil {
				ErrorHandler(w, "Invalid category ID", http.StatusBadRequest)
				return
			}
			categoryID = &catID // Присваиваем значение указателю
		}

		if title == "" || content == "" {
			checker = "All fields are required"
			http.Redirect(w, r, "/create-post", http.StatusSeeOther)
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
		var savePath string
		file, handler, err := r.FormFile("image")
		if err != nil && err != http.ErrMissingFile {
			http.Error(w, "Error retrieving file", http.StatusBadRequest)
			return
		}

		if file != nil {
			defer file.Close()
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
		err = database.CreatePost(db, title, content, userID, categoryID, createdAt, savePath)
		if err != nil {
			ErrorHandler(w, "Error saving post to database", http.StatusInternalServerError)
			return
		}

		// Перенаправляем на главную страницу или список постов
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
