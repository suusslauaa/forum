package handlers

import (
	"github.com/gofrs/uuid"
	"net/http"
)

func GetSessionID(w http.ResponseWriter, r *http.Request) (*http.Cookie, error) {
	sessionID, err := r.Cookie("session_id")
	if err != nil {
		sessionID = &http.Cookie{
			Name:  "session_id",
			Value: uuid.Must(uuid.NewV4()).String(),
		}
		http.SetCookie(w, sessionID)
	}
	return sessionID, err
}
