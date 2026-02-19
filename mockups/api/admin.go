package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type AdminUser struct {
	ID            int64     `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	AvatarURL     *string   `json:"avatar_url,omitempty"`
	OAuthProvider string    `json:"oauth_provider"`
	IsAdmin       bool      `json:"is_admin"`
	IsApproved    bool      `json:"is_approved"`
	ItemCount     int       `json:"item_count"`
	OutfitCount   int       `json:"outfit_count"`
	CreatedAt     time.Time `json:"created_at"`
	LastActive    time.Time `json:"last_active"`
}

func handleAdminGetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := dbPool.Query(context.Background(), `
		SELECT u.id, u.email, u.name, u.avatar_url, u.oauth_provider, u.is_admin, u.is_approved, u.created_at,
		       COALESCE(i.item_count, 0) as item_count,
		       COALESCE(o.outfit_count, 0) as outfit_count,
		       COALESCE(u.updated_at, u.created_at) as last_active
		FROM users u
		LEFT JOIN (SELECT user_id, COUNT(*) as item_count FROM items GROUP BY user_id) i ON u.id = i.user_id
		LEFT JOIN (SELECT user_id, COUNT(*) as outfit_count FROM outfits GROUP BY user_id) o ON u.id = o.user_id
		ORDER BY u.created_at DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	defer rows.Close()

	var users []AdminUser
	for rows.Next() {
		var u AdminUser
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.OAuthProvider, &u.IsAdmin, &u.IsApproved, &u.CreatedAt, &u.ItemCount, &u.OutfitCount, &u.LastActive); err != nil {
			continue
		}
		users = append(users, u)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func handleAdminGetStats(w http.ResponseWriter, r *http.Request) {
	stats := make(map[string]interface{})

	var totalUsers, pendingUsers, adminUsers, totalItems, totalOutfits int64
	dbPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	dbPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE is_approved = false").Scan(&pendingUsers)
	dbPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE is_admin = true").Scan(&adminUsers)
	dbPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM items").Scan(&totalItems)
	dbPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM outfits").Scan(&totalOutfits)

	stats["total_users"] = totalUsers
	stats["pending_users"] = pendingUsers
	stats["admin_users"] = adminUsers
	stats["total_items"] = totalItems
	stats["total_outfits"] = totalOutfits

	var lastSignup *time.Time
	dbPool.QueryRow(context.Background(), "SELECT MAX(created_at) FROM users").Scan(&lastSignup)
	if lastSignup != nil {
		stats["last_signup"] = lastSignup
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func handleAdminApproveUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	result, err := dbPool.Exec(context.Background(),
		"UPDATE users SET is_approved = true, updated_at = NOW() WHERE id = $1", userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "approved"})
}

func handleAdminSetAdmin(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	var body struct {
		IsAdmin bool `json:"is_admin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	currentUser := getUserFromRequest(r)
	if currentUser != nil {
		currentUserID := currentUser.ID
		var targetID int64
		dbPool.QueryRow(context.Background(), "SELECT id FROM users WHERE id = $1", userID).Scan(&targetID)
		if targetID == currentUserID && !body.IsAdmin {
			writeError(w, http.StatusBadRequest, "Cannot remove your own admin status")
			return
		}
	}

	result, err := dbPool.Exec(context.Background(),
		"UPDATE users SET is_admin = $1, updated_at = NOW() WHERE id = $2", body.IsAdmin, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"is_admin": body.IsAdmin})
}

func handleAdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	currentUser := getUserFromRequest(r)
	if currentUser != nil {
		var targetID int64
		dbPool.QueryRow(context.Background(), "SELECT id FROM users WHERE id = $1", userID).Scan(&targetID)
		if targetID == currentUser.ID {
			writeError(w, http.StatusBadRequest, "Cannot delete your own account")
			return
		}
	}

	tx, err := dbPool.Begin(context.Background())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "DELETE FROM outfit_items WHERE outfit_id IN (SELECT id FROM outfits WHERE user_id = $1)", userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	_, err = tx.Exec(context.Background(), "DELETE FROM wear_history WHERE item_id IN (SELECT id FROM items WHERE user_id = $1)", userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	_, err = tx.Exec(context.Background(), "DELETE FROM outfits WHERE user_id = $1", userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	_, err = tx.Exec(context.Background(), "DELETE FROM items WHERE user_id = $1", userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	result, err := tx.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "User not found")
		return
	}

	if err := tx.Commit(context.Background()); err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}

func handleAdminRejectUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "Missing user ID")
		return
	}

	currentUser := getUserFromRequest(r)
	if currentUser != nil {
		var targetID int64
		dbPool.QueryRow(context.Background(), "SELECT id FROM users WHERE id = $1", userID).Scan(&targetID)
		if targetID == currentUser.ID {
			writeError(w, http.StatusBadRequest, "Cannot reject your own account")
			return
		}
	}

	result, err := dbPool.Exec(context.Background(), "DELETE FROM users WHERE id = $1 AND is_approved = false", userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	if result.RowsAffected() == 0 {
		writeError(w, http.StatusNotFound, "User not found or already approved")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "rejected"})
}

func handleAdminApproveAll(w http.ResponseWriter, r *http.Request) {
	result, err := dbPool.Exec(context.Background(),
		"UPDATE users SET is_approved = true, updated_at = NOW() WHERE is_approved = false")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"approved": result.RowsAffected()})
}
