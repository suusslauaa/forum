package handlers

import (
	"github.com/gofrs/uuid"
	"net/http"
	"time"
)

func GetSessionID(w http.ResponseWriter, r *http.Request) (*http.Cookie, error) {
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		sessionID = &http.Cookie{
			Name:     "session_id",
			Value:    uuid.Must(uuid.NewV4()).String(),
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
			MaxAge:   3600,                      // Сессия живёт 1 час
			Expires:  time.Now().Add(time.Hour), // Явное время истечения
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, sessionID)
	}
	return sessionID, err
}
