package handlers

import (
	"fmt"
	"forum/database"
	"forum/templates"
	"html/template"
	"net/http"
	"strconv"
)

func ShowPromotionFormHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что пользователь авторизован
	sessionID, err := GetSessionID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	// Получаем данные из сессии
	username, loggedIn := store[sessionID.Value]
	if !loggedIn {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	userID, _ := id[sessionID.Value]
	db, err := database.InitDB()
	if err != nil {
		ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
		return
	}

	defer db.Close()
	var role string
	role, err = GetUserRole(db, userID)
	if err != nil {
		ErrorHandler(w, "Error fetching user role", http.StatusInternalServerError)
		return
	}
	Moders := false
	if role == "moder" || role == "admin" {
		Moders = true
	}

	admin := false
	if role == "admin" {
		admin = true
	}

	// Загружаем шаблон для формы подачи заявки
	tmpl, err := template.ParseFS(templates.Files, "formoder.html")
	if err != nil {
		ErrorHandler(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Передаем данные в шаблон (например, имя пользователя)
	data := map[string]interface{}{
		"Username": username,
		"Moders":   Moders,
		"Admin":    admin,
	}

	// Отображаем шаблон
	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Println(err)
	}
}

func SubmitPromotionRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что пользователь авторизован
	sessionID, err := GetSessionID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	// Получаем данные из сессии
	username, loggedIn := store[sessionID.Value]
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
	userID, _ := id[sessionID.Value]
	// Получаем ID пользователя из базы данных
	db, err := database.InitDB()
	if err != nil {
		ErrorHandler(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	var role string
	role, err = GetUserRole(db, userID)
	if err != nil {
		ErrorHandler(w, "Error fetching user role", http.StatusInternalServerError)
		return
	}
	Moders := false
	if role == "moder" || role == "admin" {
		Moders = true
	}

	admin := false
	if role == "admin" {
		admin = true
	}

	// Сохраняем заявку на повышение в базе данных
	_, err = db.Exec(`INSERT INTO promotion_requests (user_id, reason, status) VALUES (?, ?, 'pending')`, userID, reason)
	if err != nil {
		ErrorHandler(w, "Ошибка при подаче заявки", http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Username": username,
		"Moders":   Moders,
		"Admin":    admin,
	}
	// Загружаем шаблон с подтверждением подачи заявки
	tmpl, err := template.ParseFS(templates.Files, "formoders.html")
	if err != nil {
		ErrorHandler(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Отображаем шаблон подтверждения
	tmpl.Execute(w, data)
}

func AdminPromotionRequestsHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что пользователь авторизован как админ
	sessionID, err := GetSessionID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	username, loggedIn := store[sessionID.Value]
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
	Moders := true
	admin := true

	requests, err := database.GetPendingPromotionRequests(db)
	if err != nil {
		ErrorHandler(w, "Error fetching requests", http.StatusInternalServerError)
		return
	}
	data := map[string]interface{}{
		"Username": username,
		"Moders":   Moders,
		"Admin":    admin,
		"Requests": requests,
	}
	// Загружаем шаблон и передаем данные
	tmpl, err := template.ParseFS(templates.Files, "admin_promotion_requests.html")
	if err != nil {
		ErrorHandler(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, data)
}

func ApproveRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем, что пользователь авторизован как админ
	sessionID, err := GetSessionID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
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
	sessionID, err := GetSessionID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
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
