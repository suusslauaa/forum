package main

import (
	"crypto/tls"
	"forum/database"
	"forum/handlers"
	tlsecurity "forum/tls"
	"golang.org/x/time/rate"
	"io/fs"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var limiter = rate.NewLimiter(1, 5)

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
	http.Handle("/", RateLimitMiddleware(http.HandlerFunc(handlers.HomeHandler)))
	http.Handle("/register", RateLimitMiddleware(http.HandlerFunc(handlers.RegisterHandler)))
	http.Handle("/login", RateLimitMiddleware(http.HandlerFunc(handlers.LoginHandler)))
	http.Handle("/logout", RateLimitMiddleware(http.HandlerFunc(handlers.LogoutHandler)))
	http.Handle("/create-post", RateLimitMiddleware(http.HandlerFunc(handlers.CreatePostHandler)))
	http.Handle("/notifications", RateLimitMiddleware(http.HandlerFunc(handlers.NotificationsHandler)))
	http.Handle("/notifications/read", RateLimitMiddleware(http.HandlerFunc(handlers.ReadNotificationsHandler)))
	http.Handle("/edit-post", RateLimitMiddleware(http.HandlerFunc(handlers.EditPostHandler)))
	http.Handle("/post", RateLimitMiddleware(http.HandlerFunc(handlers.PostHandler)))
	http.Handle("/my-posts", RateLimitMiddleware(http.HandlerFunc(handlers.UserPostHandler)))
	http.Handle("/liked-posts", RateLimitMiddleware(http.HandlerFunc(handlers.LikePostHandler)))
	http.Handle("/comment", RateLimitMiddleware(http.HandlerFunc(handlers.CommentHandler)))
	http.Handle("/gomoder", RateLimitMiddleware(http.HandlerFunc(handlers.ShowPromotionFormHandler)))
	http.Handle("/submit-promotion-request", RateLimitMiddleware(http.HandlerFunc(handlers.SubmitPromotionRequestHandler)))
	http.Handle("/approve-request", RateLimitMiddleware(http.HandlerFunc(handlers.ApproveRequestHandler)))
	http.Handle("/deny-request", RateLimitMiddleware(http.HandlerFunc(handlers.DenyRequestHandler)))
	http.Handle("/requests", RateLimitMiddleware(http.HandlerFunc(handlers.AdminPromotionRequestsHandler)))
	http.Handle("/google/login", RateLimitMiddleware(http.HandlerFunc(handlers.GoogleLogin)))
	http.Handle("/callback", RateLimitMiddleware(http.HandlerFunc(handlers.GoogleCallback)))
	http.Handle("/github/login", RateLimitMiddleware(http.HandlerFunc(handlers.GitHubLogin)))
	http.Handle("/github/callback", RateLimitMiddleware(http.HandlerFunc(handlers.GitHubCallback)))
	http.Handle("/activity-page", RateLimitMiddleware(http.HandlerFunc(handlers.GetUserActivity)))
	http.Handle("/reports", RateLimitMiddleware(http.HandlerFunc(handlers.ReportsHandler)))
	http.Handle("/categories", RateLimitMiddleware(http.HandlerFunc(handlers.CategoryHandler)))
	http.Handle("/users", RateLimitMiddleware(http.HandlerFunc(handlers.UserListHandler)))
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

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			handlers.ErrorHandler(w, "Too many requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
