package handlers

import (
	"net/http"
)

// LogoutHandler обрабатывает выход пользователя
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем session_id из cookies
	sessionID, err := GetSessionID(w, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}

	// Удаляем сессию
	delete(store, sessionID.Value)

	// Удаляем cookie с session_id
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Перенаправляем на главную страницу
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
