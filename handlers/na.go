package handlers

import (
	"html/template"
	"net/http"
)

func NotificationsHandler(w http.ResponseWriter, r *http.Request) {
	var loggedIn bool
	var UserID int
	var username string
	// Загружаем шаблон и передаем данные
	tmpl, err := template.ParseFiles("./templates/notifications.html")
	if err != nil {
		ErrorHandler(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Создаем данные для шаблона, включая информацию о пользователе
	data := struct {
		LoggedIn bool
		ID       int
		Username string
	}{
		LoggedIn: loggedIn,
		ID:       UserID,
		Username: username,
	}

	tmpl.Execute(w, data)
}
