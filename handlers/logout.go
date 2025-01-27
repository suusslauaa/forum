package handlers

import (
	"net/http"
)

// LogoutHandler обрабатывает выход пользователя
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Получаем session_id из cookies
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Ошибка сессии", http.StatusInternalServerError)
		return
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
