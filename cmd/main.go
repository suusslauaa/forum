package main

import (
	"forum/database"
	"forum/handlers"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Открытие подключения к базе данных
	db, err := database.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// Создание категорий и постов (если нужно)
	database.CreateCategory(db, "General")
	database.CreateCategory(db, "Technology")
	//err = database.CreateUser(db, "user@example.com", "username123", "password123")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//err = database.CreatePost(db, "First Post", "This is the content of the first post.", 1, 1, time.Now().Format("2006-01-02 15:04:05"))
	//if err != nil {
	//	print("sosal?")
	//	log.Fatal(err) // Логируем и завершаем программу в случае ошибки
	//}

	// Настройка обработчиков HTTP
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/create-post", handlers.CreatePostHandler)
	http.HandleFunc("/notifications", handlers.NotificationsHandler)
	http.HandleFunc("/activity-page", handlers.ActivityPageHandler)
	http.HandleFunc("/edit-post", handlers.EditPostHandler)
	http.HandleFunc("/post", handlers.PostHandler)
	http.HandleFunc("/my-posts", handlers.UserPostHandler)
	http.HandleFunc("/liked-posts", handlers.LikePostHandler)
	http.HandleFunc("/comment", handlers.CommentHandler)
	http.HandleFunc("/gomoder", handlers.ShowPromotionFormHandler)
	http.HandleFunc("/submit-promotion-request", handlers.SubmitPromotionRequestHandler)
	http.HandleFunc("/approve-request", handlers.ApproveRequestHandler)
	http.HandleFunc("/deny-request", handlers.DenyRequestHandler)
	http.HandleFunc("/requests", handlers.AdminPromotionRequestsHandler)
	http.HandleFunc("/google/login", handlers.GoogleLogin)
	http.HandleFunc("/callback", handlers.GoogleCallback)
	http.HandleFunc("/github/login", handlers.GitHubLogin)
	http.HandleFunc("/github/callback", handlers.GitHubCallback)

	// Запуск сервера
	log.Println("Сервер запущен на http://localhost:4000")
	log.Fatal(http.ListenAndServe(":4000", nil))
}
