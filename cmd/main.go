package main

import (
	"crypto/tls"
	"forum/database"
	"forum/handlers"
	tlsecurity "forum/tls"
	"io/fs"
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
	// err = database.CreateUser(db, "user@example.com", "username123", "password123", "admin")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Настройка обработчиков HTTP
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	http.HandleFunc("/", handlers.HomeHandler)
	http.HandleFunc("/register", handlers.RegisterHandler)
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)
	http.HandleFunc("/create-post", handlers.CreatePostHandler)
	http.HandleFunc("/notifications", handlers.NotificationsHandler)
	http.HandleFunc("/notifications/read", handlers.ReadNotificationsHandler)
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
	http.HandleFunc("/activity-page", handlers.GetUserActivity)
	http.HandleFunc("/reports", handlers.ReportsHandler)
	http.HandleFunc("/categories", handlers.CategoryHandler)
	http.HandleFunc("/users", handlers.UserListHandler)
	certData, err := fs.ReadFile(tlsecurity.Pems, "cert.pem")
	if err != nil {
		log.Fatal("Failed to read TLS certificate", "error", err)
	}
	keyData, err := fs.ReadFile(tlsecurity.Pems, "key.pem")
	if err != nil {
		log.Fatal("Failed to read TLS key", "error", err)
	}
	cert, err := tls.X509KeyPair(certData, keyData)
	if err != nil {
		log.Fatal("Ошибка создания сертификата:", err)
	}

	server := &http.Server{
		Addr: ":4000",
		TLSConfig: &tls.Config{
			Certificates:     []tls.Certificate{cert},
			CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
			MinVersion:       tls.VersionTLS12,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
		},
	}
	// Запуск сервера
	log.Println("Сервер запущен на https://localhost:4000")
	log.Fatal(server.ListenAndServeTLS("", ""))
}
