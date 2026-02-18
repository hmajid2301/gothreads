package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool

func initDB() error {
	var err error
	connString := getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:15433/gothreads?sslmode=disable")

	dbPool, err = pgxpool.New(context.Background(), connString)
	if err != nil {
		return err
	}

	// Test connection
	if err := dbPool.Ping(context.Background()); err != nil {
		return err
	}

	log.Println("📊 Database connected:", connString)

	// Ensure default user exists (mockup uses hardcoded user_id=1)
	_, err = dbPool.Exec(context.Background(),
		`INSERT INTO users (id, email, name) VALUES (1, 'local@gothreads.dev', 'Local User') ON CONFLICT (id) DO NOTHING`)
	if err != nil {
		log.Printf("⚠️  Could not ensure default user: %v", err)
	}

	// Ensure max_wears column exists
	_, err = dbPool.Exec(context.Background(),
		`ALTER TABLE items ADD COLUMN IF NOT EXISTS max_wears INT DEFAULT 5`)
	if err != nil {
		log.Printf("⚠️  Could not add max_wears column: %v", err)
	}

	// Ensure ratings column exists (for 3-tier rating system)
	_, err = dbPool.Exec(context.Background(),
		`ALTER TABLE outfits ADD COLUMN IF NOT EXISTS ratings JSONB`)
	if err != nil {
		log.Printf("⚠️  Could not add ratings column: %v", err)
	}

	return nil
}

// GET /api/items - List all items for user
func handleListItems(w http.ResponseWriter, r *http.Request) {
	userID := int64(1) // TODO: Get from auth

	rows, err := dbPool.Query(context.Background(),
		"SELECT id, name, description, category, price, brand, color, image_url, wear_count, max_wears, ai_analysis, tags FROM items WHERE user_id = $1 ORDER BY created_at DESC",
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
		var description, brand, color, aiAnalysis *string
		var wearCount, maxWears int32
		var tags []string

		if err := rows.Scan(&id, &name, &description, &category, &price, &brand, &color, &imageURL, &wearCount, &maxWears, &aiAnalysis, &tags); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		item := map[string]interface{}{
			"id":         id,
			"name":       name,
			"category":   category,
			"price":      price,
			"image_url":  imageURL,
			"wear_count": wearCount,
			"max_wears":  maxWears,
		}

		if description != nil {
			item["description"] = *description
		}
		if brand != nil {
			item["brand"] = *brand
		}
		if color != nil {
			item["color"] = *color
		}
		if aiAnalysis != nil {
			item["ai_analysis"] = *aiAnalysis
		}
		if tags != nil {
			item["tags"] = tags
		}

		items = append(items, item)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
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
	var description, brand, color, aiAnalysis, season, notes *string
	var wearCount, maxWears int32
	var tags []string

	err = dbPool.QueryRow(context.Background(),
		`SELECT id, name, description, category, price, brand, color, season, image_url, notes, wear_count, max_wears, ai_analysis, tags
		 FROM items WHERE id = $1 AND user_id = $2`,
		itemID, userID).Scan(&id, &name, &description, &category, &price, &brand, &color, &season, &imageURL, &notes, &wearCount, &maxWears, &aiAnalysis, &tags)

	if err != nil {
		log.Printf("Error getting item %d: %v", itemID, err)
		writeError(w, http.StatusNotFound, "Item not found")
		return
	}

	item := map[string]interface{}{
		"id":         id,
		"name":       name,
		"category":   category,
		"price":      price,
		"image_url":  imageURL,
		"wear_count": wearCount,
		"max_wears":  maxWears,
	}

	if description != nil {
		item["description"] = *description
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
		item["ai_analysis"] = *aiAnalysis
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
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Category    string   `json:"category"`
		Price       float64  `json:"price"`
		Brand       string   `json:"brand"`
		Color       string   `json:"color"`
		Season      string   `json:"season"`
		ImageURL    string   `json:"image_url"`
		Notes       string   `json:"notes"`
		AIAnalysis  string   `json:"ai_analysis"`
		Tags        []string `json:"tags"`
		MaxWears    int      `json:"max_wears"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	maxWears := req.MaxWears
	if maxWears == 0 {
		switch req.Category {
		case "Tops":
			maxWears = 2
		case "Bottoms":
			maxWears = 5
		case "Outerwear":
			maxWears = 10
		case "Shoes":
			maxWears = 50
		case "Dresses", "One-Pieces":
			maxWears = 2
		default:
			maxWears = 5
		}
	}

	var itemID int64
	err := dbPool.QueryRow(context.Background(),
		`INSERT INTO items (user_id, name, description, category, price, brand, color, season, image_url, notes, ai_analysis, tags, max_wears)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		 RETURNING id`,
		userID, req.Name, nullString(req.Description), req.Category, req.Price, nullString(req.Brand), nullString(req.Color),
		nullString(req.Season), req.ImageURL, nullString(req.Notes), nullString(req.AIAnalysis), req.Tags, maxWears).Scan(&itemID)

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
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Category    string   `json:"category"`
		Price       float64  `json:"price"`
		Brand       string   `json:"brand"`
		Color       string   `json:"color"`
		Season      string   `json:"season"`
		Notes       string   `json:"notes"`
		AIAnalysis  string   `json:"ai_analysis"`
		Tags        []string `json:"tags"`
		MaxWears    *int     `json:"max_wears"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	_, err = dbPool.Exec(context.Background(),
		`UPDATE items SET name = $1, description = $2, category = $3, price = $4, brand = $5, color = $6, season = $7,
		                  notes = $8, ai_analysis = $9, tags = $10, max_wears = COALESCE($11, max_wears), updated_at = NOW()
		 WHERE id = $12 AND user_id = $13`,
		req.Name, nullString(req.Description), req.Category, req.Price, nullString(req.Brand), nullString(req.Color),
		nullString(req.Season), nullString(req.Notes), nullString(req.AIAnalysis), req.Tags, req.MaxWears, itemID, userID)

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

// ============================================================================
// OUTFIT CRUD
// ============================================================================

// GET /api/outfits - List all outfits for user
func handleListOutfits(w http.ResponseWriter, r *http.Request) {
	userID := int64(1) // TODO: Get from auth

	rows, err := dbPool.Query(context.Background(),
		`SELECT id, name, notes, wear_count, rating, ratings, body_image_url, created_at FROM outfits WHERE user_id = $1 ORDER BY created_at DESC`,
		userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	defer rows.Close()

	var outfits []map[string]interface{}
	for rows.Next() {
		var id int64
		var name string
		var notes, bodyImageURL *string
		var wearCount int32
		var rating *float64
		var ratings []byte // JSONB raw bytes
		var createdAt time.Time

		if err := rows.Scan(&id, &name, &notes, &wearCount, &rating, &ratings, &bodyImageURL, &createdAt); err != nil {
			log.Printf("Error scanning outfit row: %v", err)
			continue
		}

		outfit := map[string]interface{}{
			"id":         id,
			"name":       name,
			"wear_count": wearCount,
			"created_at": createdAt,
		}
		if notes != nil {
			outfit["notes"] = *notes
		}
		if rating != nil {
			outfit["rating"] = *rating
		}
		if ratings != nil {
			var r map[string]interface{}
			if err := json.Unmarshal(ratings, &r); err == nil {
				outfit["ratings"] = r
			}
		}
		if bodyImageURL != nil {
			outfit["body_image_url"] = *bodyImageURL
		}

		// Fetch outfit items
		itemRows, err := dbPool.Query(context.Background(),
			`SELECT item_id, position_x, position_y, position_z, scale FROM outfit_items WHERE outfit_id = $1`,
			id)
		if err == nil {
			var itemIDs []int64
			var positions []map[string]interface{}
			for itemRows.Next() {
				var itemID int64
				var px, py, pz int32
				var scale float64
				if err := itemRows.Scan(&itemID, &px, &py, &pz, &scale); err == nil {
					itemIDs = append(itemIDs, itemID)
					positions = append(positions, map[string]interface{}{
						"id": itemID, "x": px, "y": py, "z": pz, "scale": scale,
					})
				}
			}
			itemRows.Close()
			outfit["item_ids"] = itemIDs
			outfit["positions"] = positions
		}

		outfits = append(outfits, outfit)
	}

	if outfits == nil {
		outfits = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(outfits)
}

// GET /api/outfits/{id} - Get single outfit
func handleGetOutfit(w http.ResponseWriter, r *http.Request) {
	userID := int64(1)
	outfitIDStr := r.PathValue("id")
	outfitID, err := strconv.ParseInt(outfitIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid outfit ID")
		return
	}

	var id int64
	var name string
	var notes, bodyImageURL *string
	var wearCount int32
	var rating *float64
	var ratings []byte
	var createdAt time.Time

	err = dbPool.QueryRow(context.Background(),
		`SELECT id, name, notes, wear_count, rating, ratings, body_image_url, created_at FROM outfits WHERE id = $1 AND user_id = $2`,
		outfitID, userID).Scan(&id, &name, &notes, &wearCount, &rating, &ratings, &bodyImageURL, &createdAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "Outfit not found")
		return
	}

	outfit := map[string]interface{}{
		"id":         id,
		"name":       name,
		"wear_count": wearCount,
		"created_at": createdAt,
	}
	if notes != nil {
		outfit["notes"] = *notes
	}
	if rating != nil {
		outfit["rating"] = *rating
	}
	if ratings != nil {
		var r map[string]interface{}
		if err := json.Unmarshal(ratings, &r); err == nil {
			outfit["ratings"] = r
		}
	}
	if bodyImageURL != nil {
		outfit["body_image_url"] = *bodyImageURL
	}

	// Fetch outfit items
	itemRows, err := dbPool.Query(context.Background(),
		`SELECT item_id, position_x, position_y, position_z, scale FROM outfit_items WHERE outfit_id = $1`,
		id)
	if err == nil {
		var itemIDs []int64
		var positions []map[string]interface{}
		for itemRows.Next() {
			var itemID int64
			var px, py, pz int32
			var scale float64
			if err := itemRows.Scan(&itemID, &px, &py, &pz, &scale); err == nil {
				itemIDs = append(itemIDs, itemID)
				positions = append(positions, map[string]interface{}{
					"id": itemID, "x": px, "y": py, "z": pz, "scale": scale,
				})
			}
		}
		itemRows.Close()
		outfit["item_ids"] = itemIDs
		outfit["positions"] = positions
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(outfit)
}

// POST /api/outfits - Create new outfit
func handleCreateOutfit(w http.ResponseWriter, r *http.Request) {
	userID := int64(1)

	var req struct {
		Name         string         `json:"name"`
		Notes        string         `json:"notes"`
		BodyImageURL string         `json:"body_image_url"`
		Ratings      map[string]int `json:"ratings"`
		Positions    []struct {
			ID    int64   `json:"id"`
			X     int     `json:"x"`
			Y     int     `json:"y"`
			Z     int     `json:"z"`
			Scale float64 `json:"scale"`
		} `json:"positions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	var outfitID int64
	err := dbPool.QueryRow(context.Background(),
		`INSERT INTO outfits (user_id, name, notes, body_image_url, ratings)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		userID, req.Name, nullString(req.Notes), nullString(req.BodyImageURL), req.Ratings).Scan(&outfitID)
	if err != nil {
		log.Printf("Error creating outfit: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create outfit")
		return
	}

	// Insert outfit items
	for _, pos := range req.Positions {
		scale := pos.Scale
		if scale == 0 {
			scale = 1.0
		}
		_, err := dbPool.Exec(context.Background(),
			`INSERT INTO outfit_items (outfit_id, item_id, position_x, position_y, position_z, scale)
			 VALUES ($1, $2, $3, $4, $5, $6)
			 ON CONFLICT (outfit_id, item_id) DO UPDATE SET position_x = $3, position_y = $4, position_z = $5, scale = $6`,
			outfitID, pos.ID, pos.X, pos.Y, pos.Z, scale)
		if err != nil {
			log.Printf("Error inserting outfit item %d: %v", pos.ID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      outfitID,
		"message": "Outfit created successfully",
	})
}

// PUT /api/outfits/{id} - Update outfit
func handleUpdateOutfit(w http.ResponseWriter, r *http.Request) {
	userID := int64(1)
	outfitIDStr := r.PathValue("id")
	outfitID, err := strconv.ParseInt(outfitIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid outfit ID")
		return
	}

	var req struct {
		Name         string         `json:"name"`
		Notes        string         `json:"notes"`
		BodyImageURL string         `json:"body_image_url"`
		Rating       *float64       `json:"rating"`
		Ratings      map[string]int `json:"ratings"`
		Positions    []struct {
			ID    int64   `json:"id"`
			X     int     `json:"x"`
			Y     int     `json:"y"`
			Z     int     `json:"z"`
			Scale float64 `json:"scale"`
		} `json:"positions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	_, err = dbPool.Exec(context.Background(),
		`UPDATE outfits SET name = $1, notes = $2, body_image_url = $3, rating = $4, ratings = $5, updated_at = NOW()
		 WHERE id = $6 AND user_id = $7`,
		req.Name, nullString(req.Notes), nullString(req.BodyImageURL), req.Rating, req.Ratings, outfitID, userID)
	if err != nil {
		log.Printf("Error updating outfit: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to update outfit")
		return
	}

	// Delete old items and re-insert
	dbPool.Exec(context.Background(), `DELETE FROM outfit_items WHERE outfit_id = $1`, outfitID)
	for _, pos := range req.Positions {
		scale := pos.Scale
		if scale == 0 {
			scale = 1.0
		}
		dbPool.Exec(context.Background(),
			`INSERT INTO outfit_items (outfit_id, item_id, position_x, position_y, position_z, scale)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			outfitID, pos.ID, pos.X, pos.Y, pos.Z, scale)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Outfit updated successfully",
	})
}

// DELETE /api/outfits/{id} - Delete outfit
func handleDeleteOutfit(w http.ResponseWriter, r *http.Request) {
	userID := int64(1)
	outfitIDStr := r.PathValue("id")
	outfitID, err := strconv.ParseInt(outfitIDStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid outfit ID")
		return
	}

	_, err = dbPool.Exec(context.Background(),
		"DELETE FROM outfits WHERE id = $1 AND user_id = $2",
		outfitID, userID)
	if err != nil {
		log.Printf("Error deleting outfit: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to delete outfit")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// CALENDAR CRUD
// ============================================================================

// GET /api/calendar - List calendar events (optional ?start=&end= date range)
func handleListCalendarEvents(w http.ResponseWriter, r *http.Request) {
	userID := int64(1)

	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")

	query := `SELECT id, event_date, event_name, weather, location, outfit_id, notes FROM calendar_events WHERE user_id = $1`
	args := []interface{}{userID}

	if startDate != "" && endDate != "" {
		query += ` AND event_date >= $2 AND event_date <= $3`
		args = append(args, startDate, endDate)
	}
	query += ` ORDER BY event_date DESC`

	pgRows, err := dbPool.Query(context.Background(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	defer pgRows.Close()

	var events []map[string]interface{}
	for pgRows.Next() {
		var id int64
		var eventDate time.Time
		var eventName, weather, location, notes *string
		var outfitID *int64

		if err := pgRows.Scan(&id, &eventDate, &eventName, &weather, &location, &outfitID, &notes); err != nil {
			log.Printf("Error scanning calendar row: %v", err)
			continue
		}

		event := map[string]interface{}{
			"id":         id,
			"event_date": eventDate.Format("2006-01-02"),
		}
		if eventName != nil {
			event["event_name"] = *eventName
		}
		if weather != nil {
			event["weather"] = *weather
		}
		if location != nil {
			event["location"] = *location
		}
		if outfitID != nil {
			event["outfit_id"] = *outfitID
		}
		if notes != nil {
			event["notes"] = *notes
		}

		events = append(events, event)
	}

	if events == nil {
		events = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

// POST /api/calendar - Upsert calendar event
func handleUpsertCalendarEvent(w http.ResponseWriter, r *http.Request) {
	userID := int64(1)

	var req struct {
		Date      string `json:"date"` // YYYY-MM-DD
		EventName string `json:"event_name"`
		Weather   string `json:"weather"`
		Location  string `json:"location"`
		OutfitID  *int64 `json:"outfit_id"`
		Notes     string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Date == "" {
		writeError(w, http.StatusBadRequest, "Date is required")
		return
	}

	var eventID int64
	err := dbPool.QueryRow(context.Background(),
		`INSERT INTO calendar_events (user_id, event_date, event_name, weather, location, outfit_id, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (user_id, event_date) DO UPDATE
		 SET event_name = $3, weather = $4, location = $5, outfit_id = $6, notes = $7, updated_at = NOW()
		 RETURNING id`,
		userID, req.Date, nullString(req.EventName), nullString(req.Weather), nullString(req.Location), req.OutfitID, nullString(req.Notes)).Scan(&eventID)

	if err != nil {
		log.Printf("Error upserting calendar event: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to save calendar event")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":      eventID,
		"message": "Calendar event saved",
	})
}

// DELETE /api/calendar/{date} - Delete calendar event by date
func handleDeleteCalendarEvent(w http.ResponseWriter, r *http.Request) {
	userID := int64(1)
	date := r.PathValue("date")

	_, err := dbPool.Exec(context.Background(),
		"DELETE FROM calendar_events WHERE user_id = $1 AND event_date = $2",
		userID, date)
	if err != nil {
		log.Printf("Error deleting calendar event: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to delete calendar event")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================================================
// WEAR HISTORY
// ============================================================================

// POST /api/wear - Log wear event (wears an outfit, increments counts)
func handleLogWear(w http.ResponseWriter, r *http.Request) {
	userID := int64(1)

	var req struct {
		OutfitID int64  `json:"outfit_id"`
		Date     string `json:"date"` // optional, defaults to today
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.OutfitID == 0 {
		writeError(w, http.StatusBadRequest, "outfit_id is required")
		return
	}

	wornAt := time.Now()
	if req.Date != "" {
		if t, err := time.Parse("2006-01-02", req.Date); err == nil {
			wornAt = t
		}
	}

	// Get all items in this outfit
	itemRows, err := dbPool.Query(context.Background(),
		`SELECT item_id FROM outfit_items WHERE outfit_id = $1`, req.OutfitID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch outfit items")
		return
	}
	defer itemRows.Close()

	var itemIDs []int64
	for itemRows.Next() {
		var itemID int64
		if err := itemRows.Scan(&itemID); err == nil {
			itemIDs = append(itemIDs, itemID)
		}
	}

	// Insert wear_history rows for each item
	for _, itemID := range itemIDs {
		dbPool.Exec(context.Background(),
			`INSERT INTO wear_history (user_id, item_id, outfit_id, worn_at) VALUES ($1, $2, $3, $4)`,
			userID, itemID, req.OutfitID, wornAt)

		// Increment item wear_count
		dbPool.Exec(context.Background(),
			`UPDATE items SET wear_count = wear_count + 1 WHERE id = $1`, itemID)
	}

	// Increment outfit wear_count
	dbPool.Exec(context.Background(),
		`UPDATE outfits SET wear_count = wear_count + 1 WHERE id = $1`, req.OutfitID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Wear logged",
		"items_worn": len(itemIDs),
	})
}

// GET /api/wear - Get wear history (optional ?item_id= or ?outfit_id=)
func handleGetWearHistory(w http.ResponseWriter, r *http.Request) {
	userID := int64(1)

	query := `SELECT id, item_id, outfit_id, worn_at FROM wear_history WHERE user_id = $1`
	args := []interface{}{userID}

	if itemIDStr := r.URL.Query().Get("item_id"); itemIDStr != "" {
		if itemID, err := strconv.ParseInt(itemIDStr, 10, 64); err == nil {
			query += ` AND item_id = $2`
			args = append(args, itemID)
		}
	} else if outfitIDStr := r.URL.Query().Get("outfit_id"); outfitIDStr != "" {
		if outfitID, err := strconv.ParseInt(outfitIDStr, 10, 64); err == nil {
			query += ` AND outfit_id = $2`
			args = append(args, outfitID)
		}
	}

	query += ` ORDER BY worn_at DESC`

	rows, err := dbPool.Query(context.Background(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var id, itemID int64
		var outfitID *int64
		var wornAt time.Time

		if err := rows.Scan(&id, &itemID, &outfitID, &wornAt); err != nil {
			log.Printf("Error scanning wear history: %v", err)
			continue
		}

		entry := map[string]interface{}{
			"id":      id,
			"item_id": itemID,
			"worn_at": wornAt.Format("2006-01-02"),
		}
		if outfitID != nil {
			entry["outfit_id"] = *outfitID
		}

		history = append(history, entry)
	}

	if history == nil {
		history = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

// ============================================================================
// AI TOOL QUERY FUNCTIONS (internal, not HTTP handlers)
// These return compact data for LLM tool calling — no image_url or ai_analysis.
// ============================================================================

func toolListItems(ctx context.Context, userID int64, category, color, season, brand string, limit int) ([]map[string]interface{}, error) {
	query := `SELECT id, name, category, color, brand, season, tags, wear_count, price FROM items WHERE user_id = $1`
	args := []interface{}{userID}
	argN := 2

	if category != "" {
		query += fmt.Sprintf(` AND category ILIKE $%d`, argN)
		args = append(args, category)
		argN++
	}
	if color != "" {
		query += fmt.Sprintf(` AND color ILIKE $%d`, argN)
		args = append(args, "%"+color+"%")
		argN++
	}
	if season != "" {
		query += fmt.Sprintf(` AND season ILIKE $%d`, argN)
		args = append(args, season)
		argN++
	}
	if brand != "" {
		query += fmt.Sprintf(` AND brand ILIKE $%d`, argN)
		args = append(args, "%"+brand+"%")
		argN++
	}

	query += ` ORDER BY created_at DESC`
	if limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, limit)
	}

	rows, err := dbPool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, category string
		var color, brand, season *string
		var tags []string
		var wearCount int32
		var price float64

		if err := rows.Scan(&id, &name, &category, &color, &brand, &season, &tags, &wearCount, &price); err != nil {
			continue
		}

		item := map[string]interface{}{
			"id":         id,
			"name":       name,
			"category":   category,
			"wear_count": wearCount,
			"price":      price,
		}
		if color != nil {
			item["color"] = *color
		}
		if brand != nil {
			item["brand"] = *brand
		}
		if season != nil {
			item["season"] = *season
		}
		if tags != nil {
			item["tags"] = tags
		}
		items = append(items, item)
	}

	if items == nil {
		items = []map[string]interface{}{}
	}
	return items, nil
}

func toolGetItem(ctx context.Context, userID int64, itemID int64) (map[string]interface{}, error) {
	var id int64
	var name, category string
	var description, brand, color, season, notes *string
	var wearCount int32
	var price float64
	var tags []string

	err := dbPool.QueryRow(ctx,
		`SELECT id, name, description, category, price, brand, color, season, notes, wear_count, tags
		 FROM items WHERE id = $1 AND user_id = $2`,
		itemID, userID).Scan(&id, &name, &description, &category, &price, &brand, &color, &season, &notes, &wearCount, &tags)
	if err != nil {
		return nil, fmt.Errorf("item not found")
	}

	item := map[string]interface{}{
		"id":         id,
		"name":       name,
		"category":   category,
		"price":      price,
		"wear_count": wearCount,
	}
	if description != nil {
		item["description"] = *description
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
	if tags != nil {
		item["tags"] = tags
	}
	return item, nil
}

func toolListOutfits(ctx context.Context, userID int64, limit int) ([]map[string]interface{}, error) {
	query := `SELECT id, name, notes, wear_count, rating, created_at FROM outfits WHERE user_id = $1 ORDER BY created_at DESC`
	if limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, limit)
	}

	rows, err := dbPool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var outfits []map[string]interface{}
	for rows.Next() {
		var id int64
		var name string
		var notes *string
		var wearCount int32
		var rating *float64
		var createdAt time.Time

		if err := rows.Scan(&id, &name, &notes, &wearCount, &rating, &createdAt); err != nil {
			continue
		}

		outfit := map[string]interface{}{
			"id":         id,
			"name":       name,
			"wear_count": wearCount,
			"created_at": createdAt.Format("2006-01-02"),
		}
		if notes != nil {
			outfit["notes"] = *notes
		}
		if rating != nil {
			outfit["rating"] = *rating
		}

		// Get item IDs
		itemRows, err := dbPool.Query(ctx,
			`SELECT oi.item_id, i.name FROM outfit_items oi JOIN items i ON i.id = oi.item_id WHERE oi.outfit_id = $1`, id)
		if err == nil {
			var itemIDs []int64
			var itemNames []string
			for itemRows.Next() {
				var iid int64
				var iname string
				if err := itemRows.Scan(&iid, &iname); err == nil {
					itemIDs = append(itemIDs, iid)
					itemNames = append(itemNames, iname)
				}
			}
			itemRows.Close()
			outfit["item_ids"] = itemIDs
			outfit["item_names"] = itemNames
		}

		outfits = append(outfits, outfit)
	}

	if outfits == nil {
		outfits = []map[string]interface{}{}
	}
	return outfits, nil
}

func toolGetOutfit(ctx context.Context, userID int64, outfitID int64) (map[string]interface{}, error) {
	var id int64
	var name string
	var notes *string
	var wearCount int32
	var rating *float64
	var createdAt time.Time

	err := dbPool.QueryRow(ctx,
		`SELECT id, name, notes, wear_count, rating, created_at FROM outfits WHERE id = $1 AND user_id = $2`,
		outfitID, userID).Scan(&id, &name, &notes, &wearCount, &rating, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("outfit not found")
	}

	outfit := map[string]interface{}{
		"id":         id,
		"name":       name,
		"wear_count": wearCount,
		"created_at": createdAt.Format("2006-01-02"),
	}
	if notes != nil {
		outfit["notes"] = *notes
	}
	if rating != nil {
		outfit["rating"] = *rating
	}

	// Get items with details
	itemRows, err := dbPool.Query(ctx,
		`SELECT i.id, i.name, i.category, i.color, i.brand, i.tags, oi.position_x, oi.position_y, oi.position_z
		 FROM outfit_items oi JOIN items i ON i.id = oi.item_id
		 WHERE oi.outfit_id = $1`, outfitID)
	if err == nil {
		var items []map[string]interface{}
		for itemRows.Next() {
			var iid int64
			var iname, icategory string
			var icolor, ibrand *string
			var itags []string
			var px, py, pz int32

			if err := itemRows.Scan(&iid, &iname, &icategory, &icolor, &ibrand, &itags, &px, &py, &pz); err == nil {
				item := map[string]interface{}{
					"id": iid, "name": iname, "category": icategory,
					"position": map[string]interface{}{"x": px, "y": py, "z": pz},
				}
				if icolor != nil {
					item["color"] = *icolor
				}
				if ibrand != nil {
					item["brand"] = *ibrand
				}
				if itags != nil {
					item["tags"] = itags
				}
				items = append(items, item)
			}
		}
		itemRows.Close()
		outfit["items"] = items
	}

	return outfit, nil
}

func toolGetWardrobeStats(ctx context.Context, userID int64) (map[string]interface{}, error) {
	stats := map[string]interface{}{}

	// Total count
	var total int64
	dbPool.QueryRow(ctx, `SELECT COUNT(*) FROM items WHERE user_id = $1`, userID).Scan(&total)
	stats["total_items"] = total

	// By category
	byCategory := map[string]int64{}
	rows, err := dbPool.Query(ctx, `SELECT category, COUNT(*) FROM items WHERE user_id = $1 GROUP BY category ORDER BY COUNT(*) DESC`, userID)
	if err == nil {
		for rows.Next() {
			var cat string
			var cnt int64
			if rows.Scan(&cat, &cnt) == nil {
				byCategory[cat] = cnt
			}
		}
		rows.Close()
	}
	stats["by_category"] = byCategory

	// By color
	byColor := map[string]int64{}
	rows, err = dbPool.Query(ctx, `SELECT color, COUNT(*) FROM items WHERE user_id = $1 AND color IS NOT NULL GROUP BY color ORDER BY COUNT(*) DESC`, userID)
	if err == nil {
		for rows.Next() {
			var col string
			var cnt int64
			if rows.Scan(&col, &cnt) == nil {
				byColor[col] = cnt
			}
		}
		rows.Close()
	}
	stats["by_color"] = byColor

	// By season
	bySeason := map[string]int64{}
	rows, err = dbPool.Query(ctx, `SELECT season, COUNT(*) FROM items WHERE user_id = $1 AND season IS NOT NULL GROUP BY season ORDER BY COUNT(*) DESC`, userID)
	if err == nil {
		for rows.Next() {
			var sea string
			var cnt int64
			if rows.Scan(&sea, &cnt) == nil {
				bySeason[sea] = cnt
			}
		}
		rows.Close()
	}
	stats["by_season"] = bySeason

	// Total outfits
	var totalOutfits int64
	dbPool.QueryRow(ctx, `SELECT COUNT(*) FROM outfits WHERE user_id = $1`, userID).Scan(&totalOutfits)
	stats["total_outfits"] = totalOutfits

	return stats, nil
}

func toolGetWearHistory(ctx context.Context, userID int64, itemID, outfitID *int64, daysBack int) ([]map[string]interface{}, error) {
	query := `SELECT wh.id, wh.item_id, i.name AS item_name, wh.outfit_id, wh.worn_at
	          FROM wear_history wh
	          JOIN items i ON i.id = wh.item_id
	          WHERE wh.user_id = $1`
	args := []interface{}{userID}
	argN := 2

	if itemID != nil {
		query += fmt.Sprintf(` AND wh.item_id = $%d`, argN)
		args = append(args, *itemID)
		argN++
	}
	if outfitID != nil {
		query += fmt.Sprintf(` AND wh.outfit_id = $%d`, argN)
		args = append(args, *outfitID)
		argN++
	}
	if daysBack > 0 {
		query += fmt.Sprintf(` AND wh.worn_at >= NOW() - INTERVAL '%d days'`, daysBack)
	}

	query += ` ORDER BY wh.worn_at DESC LIMIT 50`

	rows, err := dbPool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []map[string]interface{}
	for rows.Next() {
		var id, iid int64
		var itemName string
		var oid *int64
		var wornAt time.Time

		if err := rows.Scan(&id, &iid, &itemName, &oid, &wornAt); err != nil {
			continue
		}

		entry := map[string]interface{}{
			"id":        id,
			"item_id":   iid,
			"item_name": itemName,
			"worn_at":   wornAt.Format("2006-01-02"),
		}
		if oid != nil {
			entry["outfit_id"] = *oid
		}
		history = append(history, entry)
	}

	if history == nil {
		history = []map[string]interface{}{}
	}
	return history, nil
}

func toolGetMostWorn(ctx context.Context, userID int64, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := dbPool.Query(ctx,
		`SELECT id, name, category, color, brand, wear_count FROM items
		 WHERE user_id = $1 AND wear_count > 0
		 ORDER BY wear_count DESC LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, category string
		var color, brand *string
		var wearCount int32

		if err := rows.Scan(&id, &name, &category, &color, &brand, &wearCount); err != nil {
			continue
		}

		item := map[string]interface{}{
			"id": id, "name": name, "category": category, "wear_count": wearCount,
		}
		if color != nil {
			item["color"] = *color
		}
		if brand != nil {
			item["brand"] = *brand
		}
		items = append(items, item)
	}

	if items == nil {
		items = []map[string]interface{}{}
	}
	return items, nil
}

func toolGetLeastWorn(ctx context.Context, userID int64, limit int) ([]map[string]interface{}, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := dbPool.Query(ctx,
		`SELECT id, name, category, color, brand, wear_count FROM items
		 WHERE user_id = $1
		 ORDER BY wear_count ASC LIMIT $2`,
		userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, category string
		var color, brand *string
		var wearCount int32

		if err := rows.Scan(&id, &name, &category, &color, &brand, &wearCount); err != nil {
			continue
		}

		item := map[string]interface{}{
			"id": id, "name": name, "category": category, "wear_count": wearCount,
		}
		if color != nil {
			item["color"] = *color
		}
		if brand != nil {
			item["brand"] = *brand
		}
		items = append(items, item)
	}

	if items == nil {
		items = []map[string]interface{}{}
	}
	return items, nil
}

func toolGetCalendarEvents(ctx context.Context, userID int64, startDate, endDate string) ([]map[string]interface{}, error) {
	query := `SELECT ce.id, ce.event_date, ce.event_name, ce.weather, ce.location, ce.outfit_id, o.name AS outfit_name
	          FROM calendar_events ce
	          LEFT JOIN outfits o ON o.id = ce.outfit_id
	          WHERE ce.user_id = $1`
	args := []interface{}{userID}
	argN := 2

	if startDate != "" {
		query += fmt.Sprintf(` AND ce.event_date >= $%d`, argN)
		args = append(args, startDate)
		argN++
	}
	if endDate != "" {
		query += fmt.Sprintf(` AND ce.event_date <= $%d`, argN)
		args = append(args, endDate)
		argN++
	}

	query += ` ORDER BY ce.event_date DESC LIMIT 50`

	rows, err := dbPool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []map[string]interface{}
	for rows.Next() {
		var id int64
		var eventDate time.Time
		var eventName, weather, location, outfitName *string
		var outfitID *int64

		if err := rows.Scan(&id, &eventDate, &eventName, &weather, &location, &outfitID, &outfitName); err != nil {
			continue
		}

		event := map[string]interface{}{
			"id":         id,
			"event_date": eventDate.Format("2006-01-02"),
		}
		if eventName != nil {
			event["event_name"] = *eventName
		}
		if weather != nil {
			event["weather"] = *weather
		}
		if location != nil {
			event["location"] = *location
		}
		if outfitID != nil {
			event["outfit_id"] = *outfitID
		}
		if outfitName != nil {
			event["outfit_name"] = *outfitName
		}
		events = append(events, event)
	}

	if events == nil {
		events = []map[string]interface{}{}
	}
	return events, nil
}

func toolSearchItems(ctx context.Context, userID int64, query string) ([]map[string]interface{}, error) {
	pattern := "%" + query + "%"

	rows, err := dbPool.Query(ctx,
		`SELECT id, name, category, color, brand, tags, wear_count, description FROM items
		 WHERE user_id = $1 AND (
		   name ILIKE $2
		   OR brand ILIKE $2
		   OR description ILIKE $2
		   OR color ILIKE $2
		 )
		 ORDER BY created_at DESC LIMIT 20`,
		userID, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, category string
		var color, brand, description *string
		var tags []string
		var wearCount int32

		if err := rows.Scan(&id, &name, &category, &color, &brand, &tags, &wearCount, &description); err != nil {
			continue
		}

		item := map[string]interface{}{
			"id": id, "name": name, "category": category, "wear_count": wearCount,
		}
		if color != nil {
			item["color"] = *color
		}
		if brand != nil {
			item["brand"] = *brand
		}
		if description != nil {
			item["description"] = *description
		}
		if tags != nil {
			item["tags"] = tags
		}
		items = append(items, item)
	}

	if items == nil {
		items = []map[string]interface{}{}
	}
	return items, nil
}
