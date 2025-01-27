package main

import (
	"forum/database"
	"forum/handlers"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"time"
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
	err = database.CreatePost(db, "First Post", "This is the content of the first post.", 1, 1, time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		log.Fatal(err) // Логируем и завершаем программу в случае ошибки
	}

	//err = database.CreatePost(db, "Second Post", "This is the content of the second post.", 2, 2, time.Now().Format("2006-01-02 15:04:05"))
	//if err != nil {
	//	print("da?")
	//	log.Fatal(err)
	//}

	// Настройка обработчиков HTTP
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/create-post", handlers.CreatePostHandler)

	// Запуск сервера
	log.Println("Сервер запущен на http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
