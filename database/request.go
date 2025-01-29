package database

import (
	"database/sql"
	"fmt"
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
	_, err := db.Exec(`UPDATE promotion_requests SET status = 'approved' WHERE id = ?`, requestID)
	return err
}

// DenyPromotionRequest обновляет статус заявки на 'denied'
func DenyPromotionRequest(db *sql.DB, requestID int) error {
	_, err := db.Exec(`UPDATE promotion_requests SET status = 'denied' WHERE id = ?`, requestID)
	return err
}
