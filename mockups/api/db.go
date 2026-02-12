package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool

func initDB() error {
	var err error
	connString := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:15432/gothreads?sslmode=disable")

	dbPool, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		return err
	}

	// Test connection
	if err := dbPool.Ping(context.Background()); err != nil {
		return err
	}

	log.Println("📊 Database connected:", connString)
	return nil
}

// GET /api/items - List all items for user
func handleListItems(w http.ResponseWriter, r *http.Request) {
	userID := int64(1) // TODO: Get from auth

	rows, err := dbPool.Query(context.Background(),
		"SELECT id, name, category, price, brand, color, image_url, wear_count, ai_analysis, tags FROM items WHERE user_id = $1 ORDER BY created_at DESC",
		userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, category, imageURL string
		var price float64
		var brand, color, aiAnalysis *string
		var wearCount int32
		var tags []string

		if err := rows.Scan(&id, &name, &category, &price, &brand, &color, &imageURL, &wearCount, &aiAnalysis, &tags); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		item := map[string]interface{}{
			"id":         id,
			"name":       name,
			"category":   category,
			"price":      price,
			"image":      imageURL,
			"wearCount":  wearCount,
		}

		if brand != nil {
			item["brand"] = *brand
		}
		if color != nil {
			item["color"] = *color
		}
		if aiAnalysis != nil {
			item["aiAnalysis"] = *aiAnalysis
		}
		if tags != nil {
			item["tags"] = tags
		}

		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"items": items,
	})
}

// GET /api/items/:id - Get single item
func handleGetItem(w http.ResponseWriter, r *http.Request) {
	userID := int64(1) // TODO: Get from auth
	itemIDStr := r.PathValue("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	var id int64
	var name, category, imageURL string
	var price float64
	var brand, color, aiAnalysis, season, notes *string
	var wearCount int32
	var tags []string

	err = dbPool.QueryRow(context.Background(),
		`SELECT id, name, category, price, brand, color, season, image_url, notes, wear_count, ai_analysis, tags
		 FROM items WHERE id = $1 AND user_id = $2`,
		itemID, userID).Scan(&id, &name, &category, &price, &brand, &color, &season, &imageURL, &notes, &wearCount, &aiAnalysis, &tags)

	if err != nil {
		log.Printf("Error getting item %d: %v", itemID, err)
		writeError(w, http.StatusNotFound, "Item not found")
		return
	}

	item := map[string]interface{}{
		"id":        id,
		"name":      name,
		"category":  category,
		"price":     price,
		"image":     imageURL,
		"wearCount": wearCount,
	}

	if brand != nil {
		item["brand"] = *brand
	}
	if color != nil {
		item["color"] = *color
	}
	if season != nil {
		item["season"] = *season
	}
	if notes != nil {
		item["notes"] = *notes
	}
	if aiAnalysis != nil {
		item["aiAnalysis"] = *aiAnalysis
	}
	if tags != nil {
		item["tags"] = tags
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

// POST /api/items - Create new item
func handleCreateItem(w http.ResponseWriter, r *http.Request) {
	userID := int64(1) // TODO: Get from auth

	var req struct {
		Name       string   `json:"name"`
		Category   string   `json:"category"`
		Price      float64  `json:"price"`
		Brand      string   `json:"brand"`
		Color      string   `json:"color"`
		Season     string   `json:"season"`
		ImageURL   string   `json:"image"`
		Notes      string   `json:"notes"`
		AIAnalysis string   `json:"aiAnalysis"`
		Tags       []string `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	var itemID int64
	err := dbPool.QueryRow(context.Background(),
		`INSERT INTO items (user_id, name, category, price, brand, color, season, image_url, notes, ai_analysis, tags)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id`,
		userID, req.Name, req.Category, req.Price, nullString(req.Brand), nullString(req.Color),
		nullString(req.Season), req.ImageURL, nullString(req.Notes), nullString(req.AIAnalysis), req.Tags).Scan(&itemID)

	if err != nil {
		log.Printf("Error creating item: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create item")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      itemID,
		"message": "Item created successfully",
	})
}

// PUT /api/items/:id - Update item
func handleUpdateItem(w http.ResponseWriter, r *http.Request) {
	userID := int64(1) // TODO: Get from auth
	itemIDStr := r.PathValue("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	var req struct {
		Name       string   `json:"name"`
		Category   string   `json:"category"`
		Price      float64  `json:"price"`
		Brand      string   `json:"brand"`
		Color      string   `json:"color"`
		Season     string   `json:"season"`
		Notes      string   `json:"notes"`
		AIAnalysis string   `json:"aiAnalysis"`
		Tags       []string `json:"tags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	_, err = dbPool.Exec(context.Background(),
		`UPDATE items SET name = $1, category = $2, price = $3, brand = $4, color = $5, season = $6,
		                  notes = $7, ai_analysis = $8, tags = $9, updated_at = NOW()
		 WHERE id = $10 AND user_id = $11`,
		req.Name, req.Category, req.Price, nullString(req.Brand), nullString(req.Color),
		nullString(req.Season), nullString(req.Notes), nullString(req.AIAnalysis), req.Tags, itemID, userID)

	if err != nil {
		log.Printf("Error updating item: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to update item")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Item updated successfully",
	})
}

// DELETE /api/items/:id - Delete item
func handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	userID := int64(1) // TODO: Get from auth
	itemIDStr := r.PathValue("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid item ID")
		return
	}

	_, err = dbPool.Exec(context.Background(),
		"DELETE FROM items WHERE id = $1 AND user_id = $2",
		itemID, userID)

	if err != nil {
		log.Printf("Error deleting item: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to delete item")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper function to convert empty strings to NULL
func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
