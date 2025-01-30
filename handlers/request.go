package handlers

import (
	"fmt"
	"forum/database"
	"html/template"
	"net/http"
	"strconv"
)

func ShowPromotionFormHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что пользователь авторизован
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Получаем данные из сессии
	username, loggedIn := store[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Загружаем шаблон для формы подачи заявки
	tmpl, err := template.ParseFiles("templates/formoder.html")
	if err != nil {
		ErrorHandler(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Передаем данные в шаблон (например, имя пользователя)
	data := struct {
		Username string
	}{
		Username: username,
	}

	// Отображаем шаблон
	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Println(err)
	}
}

func SubmitPromotionRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что пользователь авторизован
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Получаем данные из сессии
	_, loggedIn := store[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Извлекаем информацию из формы
	reason := r.FormValue("reason")
	if reason == "" {
		ErrorHandler(w, "Причина не может быть пустой", http.StatusBadRequest)
		return
	}

	// Получаем ID пользователя из базы данных
	db, err := database.InitDB()
	if err != nil {
		ErrorHandler(w, "Ошибка подключения к базе данных", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	userID := id[sessionID.Value]

	// Сохраняем заявку на повышение в базе данных
	_, err = db.Exec(`INSERT INTO promotion_requests (user_id, reason, status) VALUES (?, ?, 'pending')`, userID, reason)
	if err != nil {
		ErrorHandler(w, "Ошибка при подаче заявки", http.StatusInternalServerError)
		return
	}

	// Загружаем шаблон с подтверждением подачи заявки
	tmpl, err := template.ParseFiles("./templates/formoders.html")
	if err != nil {
		ErrorHandler(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отображаем шаблон подтверждения
	tmpl.Execute(w, nil)
}

func AdminPromotionRequestsHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что пользователь авторизован как админ
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

	// Извлекаем заявки из базы данных
	db, err := database.InitDB()
	if err != nil {
		ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	role, err := GetUserRole(db, UserID)
	if err != nil {
		ErrorHandler(w, "Error fetching user role", http.StatusInternalServerError)
		return
	}
	if role != "admin" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	requests, err := database.GetPendingPromotionRequests(db)
	if err != nil {
		ErrorHandler(w, "Error fetching requests", http.StatusInternalServerError)
		return
	}

	// Загружаем шаблон и передаем данные
	tmpl, err := template.ParseFiles("./templates/admin_promotion_requests.html")
	if err != nil {
		ErrorHandler(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, struct {
		Requests []database.PromotionRequest
	}{
		Requests: requests,
	})
}

func ApproveRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что пользователь авторизован как админ
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

	// Извлекаем ID заявки
	requestID, err := strconv.Atoi(r.FormValue("request_id"))
	if err != nil {
		ErrorHandler(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	// Обрабатываем заявку
	db, err := database.InitDB()
	if err != nil {
		ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	role, err := GetUserRole(db, UserID)
	if err != nil {
		ErrorHandler(w, "Error fetching user role", http.StatusInternalServerError)
		return
	}
	if role != "admin" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	err = database.ApprovePromotionRequest(db, requestID)
	if err != nil {
		ErrorHandler(w, "Error approving request", http.StatusInternalServerError)
		return
	}

	// Перенаправляем обратно на страницу заявок
	http.Redirect(w, r, "/admin-promotion-requests", http.StatusFound)
}

func DenyRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что пользователь авторизован как админ
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

	// Извлекаем ID заявки
	requestID, err := strconv.Atoi(r.FormValue("request_id"))
	if err != nil {
		ErrorHandler(w, "Invalid request ID", http.StatusBadRequest)
		return
	}

	// Обрабатываем заявку
	db, err := database.InitDB()
	if err != nil {
		ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	role, err := GetUserRole(db, UserID)
	if err != nil {
		ErrorHandler(w, "Error fetching user role", http.StatusInternalServerError)
		return
	}
	if role != "admin" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	err = database.DenyPromotionRequest(db, requestID)
	if err != nil {
		ErrorHandler(w, "Error denying request", http.StatusInternalServerError)
		return
	}

	// Перенаправляем обратно на страницу заявок
	http.Redirect(w, r, "/admin-promotion-requests", http.StatusFound)
}
