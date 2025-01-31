package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type PromotionRequest struct {
	ID       int
	UserID   int
	Username string
	Reason   string
	Created  time.Time
	Status   string
}

// GetPendingPromotionRequests получает все заявки на повышение с статусом 'pending'
func GetPendingPromotionRequests(db *sql.DB) ([]PromotionRequest, error) {
	rows, err := db.Query(`
		SELECT pr.id, pr.user_id, u.username, pr.reason, pr.request_date, pr.status 
		FROM promotion_requests pr
		JOIN users u ON pr.user_id = u.id
		WHERE pr.status = 'pending'`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pending promotion requests: %v", err)
	}
	defer rows.Close()

	var requests []PromotionRequest
	for rows.Next() {
		var req PromotionRequest
		err := rows.Scan(&req.ID, &req.UserID, &req.Username, &req.Reason, &req.Created, &req.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// ApprovePromotionRequest обновляет статус заявки на 'approved'
func ApprovePromotionRequest(db *sql.DB, requestID int) error {
	// Получаем user_id из таблицы promotion_requests
	var userID int
	err := db.QueryRow(`SELECT user_id FROM promotion_requests WHERE id = ?`, requestID).Scan(&userID)
	if err != nil {
		return fmt.Errorf("не удалось получить user_id: %v", err)
	}

	// Обновляем статус заявки на 'approved'
	_, err = db.Exec(`UPDATE promotion_requests SET status = 'approved' WHERE id = ?`, requestID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении статуса заявки: %v", err)
	}

	// Меняем роль пользователя (например, на 'moderator')
	newRole := "moderator"
	_, err = db.Exec(`UPDATE users SET role = ? WHERE id = ?`, newRole, userID)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении роли пользователя: %v", err)
	}

	log.Printf("Заявка %d одобрена, роль пользователя %d изменена на %s", requestID, userID, newRole)
	return nil
}

// DenyPromotionRequest обновляет статус заявки на 'denied'
func DenyPromotionRequest(db *sql.DB, requestID int) error {
	_, err := db.Exec(`UPDATE promotion_requests SET status = 'denied' WHERE id = ?`, requestID)
	return err
}

func ReportPost(db *sql.DB, postID, userID int) error {
	// Проверяем, есть ли уже репорт от этого пользователя на этот пост
	var existingStatus string
	err := db.QueryRow("SELECT status FROM reports WHERE post_id = ? AND reported_by = ?", postID, userID).Scan(&existingStatus)

	if err == sql.ErrNoRows {
		// Если репорта нет, создаем новый со статусом "open"
		_, err = db.Exec("INSERT INTO reports (post_id, reported_by, status) VALUES (?, ?, 'open')", postID, userID)
		if err != nil {
			return err
		}
	} else if err == nil {
		// Если репорт уже есть, обновляем статус на "open"
		_, err = db.Exec("UPDATE reports SET status = 'open' WHERE post_id = ? AND reported_by = ?", postID, userID)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}

func DeletePostReport(db *sql.DB, postID int) error {
	// Обновляем статус репорта, если он существует
	_, err := db.Exec("UPDATE reports SET status = 'none' WHERE post_id = ?", postID)
	if err != nil {
		return err
	}

	return nil
}
