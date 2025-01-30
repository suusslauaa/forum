package handlers

import (
	"forum/database"
	"forum/templates"
	"html/template"
	"log"
	"net/http"
)

func CategoryHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем сессию пользователя
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	_, loggedIn := store[sessionID.Value]
	UserID := id[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Подключаемся к базе данных
	db, err := database.InitDB()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		log.Println("Ошибка подключения к БД:", err)
		return
	}
	defer db.Close()

	// Проверяем роль пользователя (только admin может видеть страницу)
	role, err := GetUserRole(db, UserID)
	if err != nil {
		http.Error(w, "Error fetching user role", http.StatusInternalServerError)
		return
	}
	if role != "admin" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Обработка POST-запросов (создание, редактирование, удаление)
	if r.Method == http.MethodPost {
		action := r.URL.Query().Get("action")
		switch action {
		case "create":
			categoryName := r.FormValue("name")
			database.CreateCategory(db, categoryName)
		case "edit":
			categoryID := r.URL.Query().Get("id")
			newName := r.FormValue("name")
			_, err := db.Exec("UPDATE categories SET name = ? WHERE id = ?", newName, categoryID)
			if err != nil {
				http.Error(w, "Error updating category", http.StatusInternalServerError)
				return
			}
		case "delete":
			categoryID := r.URL.Query().Get("id")
			_, err := db.Exec("UPDATE posts SET category_id = NULL WHERE category_id = ?", categoryID)
			if err != nil {
				http.Error(w, "Error updating posts category", http.StatusInternalServerError)
				return
			}

			// Удаляем категорию
			_, err = db.Exec("DELETE FROM categories WHERE id = ?", categoryID)
			if err != nil {
				http.Error(w, "Error deleting category", http.StatusInternalServerError)
				return
			}
		default:
			http.Error(w, "Invalid action", http.StatusBadRequest)
			return
		}

		// После изменения перенаправляем обратно на страницу категорий
		http.Redirect(w, r, "/categories", http.StatusSeeOther)
		return
	}

	// Получаем список категорий
	rows, err := db.Query("SELECT id, name FROM categories")
	if err != nil {
		http.Error(w, "Error fetching categories", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []struct {
		ID   int
		Name string
	}

	for rows.Next() {
		var category struct {
			ID   int
			Name string
		}
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			http.Error(w, "Error scanning categories", http.StatusInternalServerError)
			return
		}
		categories = append(categories, category)
	}

	// Загружаем HTML-шаблон
	tmpl, err := template.ParseFS(templates.Files, "categories.html")
	if err != nil {
		http.Error(w, "Template parsing error", http.StatusInternalServerError)
		log.Println("Ошибка загрузки шаблона:", err)
		return
	}

	// Отправляем данные в шаблон
	err = tmpl.Execute(w, categories)
	if err != nil {
		http.Error(w, "Template rendering error", http.StatusInternalServerError)
		log.Println("Ошибка рендеринга:", err)
	}
}
