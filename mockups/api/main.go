// Throwaway PoC - Ollama API proxy for mockups
// Usage: go run main.go
// Requires: Ollama running on localhost:11434 with llava model

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultOllamaURL   = "http://localhost:11434"
	defaultVisionModel = "llava:7b"
	defaultTextModel   = "llama3.2:3b"
	defaultRembgURL    = "http://localhost:5000"
)

var (
	ollamaURL   = getEnv("OLLAMA_URL", defaultOllamaURL)
	visionModel = getEnv("OLLAMA_VISION_MODEL", defaultVisionModel)
	textModel   = getEnv("OLLAMA_TEXT_MODEL", defaultTextModel)
	rembgURL    = getEnv("REMBG_URL", defaultRembgURL)
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// OllamaRequest is the request format for Ollama API
type OllamaRequest struct {
	Model  string   `json:"model"`
	Prompt string   `json:"prompt"`
	Images []string `json:"images,omitempty"` // base64 encoded images
	Stream bool     `json:"stream"`
}

// OllamaResponse is the response format from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// AnalyzeRequest is what the frontend sends
type AnalyzeRequest struct {
	Image  string `json:"image"`  // base64 encoded image (with or without data URI prefix)
	Prompt string `json:"prompt"` // optional custom prompt
	Model  string `json:"model"`  // optional model override (e.g., "llava:13b")
}

// AnalyzeResponse is what we send back to frontend
type AnalyzeResponse struct {
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Color       string   `json:"color"`
	Brand       string   `json:"brand"`
	Tags        []string `json:"tags"`
	RawResponse string   `json:"raw_response,omitempty"`
}

// TagsRequest for generating tags from description
type TagsRequest struct {
	Description string `json:"description"`
}

// TagsResponse returns generated tags
type TagsResponse struct {
	Tags []string `json:"tags"`
}

// RemoveBgRequest for background removal
type RemoveBgRequest struct {
	Image string `json:"image"` // base64 encoded image
}

// RemoveBgResponse returns image with background removed
type RemoveBgResponse struct {
	Image string `json:"image"` // base64 encoded image with transparent background
}

// SuggestOutfitRequest for AI outfit recommendations
type SuggestOutfitRequest struct {
	Occasion string                 `json:"occasion"` // casual, formal, sporty, business, date, etc.
	Weather  string                 `json:"weather"`  // hot, cold, mild, rainy, etc.
	Items    []ClothingItemForMatch `json:"items"`    // available wardrobe items
}

// SuggestOutfitResponse returns suggested outfit
type SuggestOutfitResponse struct {
	SelectedItems []int    `json:"selected_items"` // IDs of items to wear
	Reasoning     string   `json:"reasoning"`      // Why these items work together
	Tips          []string `json:"tips"`           // Style tips
	RawResponse   string   `json:"raw_response,omitempty"`
}

// ColorMatchRequest for color coordination analysis
type ColorMatchRequest struct {
	Items []ClothingItemForMatch `json:"items"` // items to analyze
}

// ColorMatchResponse returns color compatibility analysis
type ColorMatchResponse struct {
	CompatibilityScore int      `json:"compatibility_score"` // 0-100
	Analysis           string   `json:"analysis"`            // Why colors work/don't work
	Suggestions        []string `json:"suggestions"`         // How to improve
	RawResponse        string   `json:"raw_response,omitempty"`
}

// StyleMatchRequest for finding matching items
type StyleMatchRequest struct {
	BaseItem ClothingItemForMatch   `json:"base_item"`      // item to match with
	Items    []ClothingItemForMatch `json:"items"`          // available items
	MaxItems int                    `json:"max_items"`      // max items to return (default 5)
}

// StyleMatchResponse returns matching items
type StyleMatchResponse struct {
	MatchingItems []MatchedItem `json:"matching_items"`
	RawResponse   string        `json:"raw_response,omitempty"`
}

type MatchedItem struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Reason   string `json:"reason"` // Why it matches
	Score    int    `json:"score"`  // 0-100 match score
}

// RateOutfitRequest for outfit feedback
type RateOutfitRequest struct {
	Items []ClothingItemForMatch `json:"items"` // items in outfit
}

// RateOutfitResponse returns outfit rating and feedback
type RateOutfitResponse struct {
	Rating      int      `json:"rating"`      // 0-10
	Feedback    string   `json:"feedback"`    // Overall feedback
	Strengths   []string `json:"strengths"`   // What works well
	Improvements []string `json:"improvements"` // What could be better
	RawResponse string   `json:"raw_response,omitempty"`
}

// WardrobeGapsRequest for analyzing wardrobe
type WardrobeGapsRequest struct {
	Items []ClothingItemForMatch `json:"items"` // all wardrobe items
}

// WardrobeGapsResponse returns wardrobe analysis
type WardrobeGapsResponse struct {
	Summary      string   `json:"summary"`       // Overall wardrobe assessment
	MissingItems []string `json:"missing_items"` // Suggested items to buy
	Strengths    []string `json:"strengths"`     // What you have well covered
	Tips         []string `json:"tips"`          // General wardrobe tips
	RawResponse  string   `json:"raw_response,omitempty"`
}

// ClothingItemForMatch simplified item info for AI matching
type ClothingItemForMatch struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Color       string   `json:"color"`
	Brand       string   `json:"brand,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Description string   `json:"description,omitempty"`
	WearCount   int      `json:"wear_count,omitempty"`
}

// DressRequest for AI-assisted clothing placement
type DressRequest struct {
	BodyImage     string                 `json:"body_image"`     // base64 body/mannequin image
	ClothingItems []ClothingItemPosition `json:"clothing_items"` // items to place
	Prompt        string                 `json:"prompt"`         // optional custom prompt
}

type ClothingItemPosition struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Image    string `json:"image"` // base64 clothing image
}

// DressResponse returns suggested positions for clothing items
type DressResponse struct {
	Positions   []ItemPosition `json:"positions"`
	Explanation string         `json:"explanation"`
	RawResponse string         `json:"raw_response,omitempty"`
}

type ItemPosition struct {
	ID int     `json:"id"`
	X  float64 `json:"x"` // percentage 0-100
	Y  float64 `json:"y"` // percentage 0-100
	Z  int     `json:"z"` // layer order
}

func main() {
	// Initialize database
	if err := initDB(); err != nil {
		log.Printf("⚠️  Database connection failed: %v", err)
		log.Printf("   Continuing without database (AI features only)...")
	}

	// Initialize S3
	if err := initS3(); err != nil {
		log.Printf("⚠️  S3 connection failed: %v", err)
		log.Printf("   Continuing without S3 (will store base64 in database)...")
	}

	mux := http.NewServeMux()

	// Serve static mockup files from parent directory
	fs := http.FileServer(http.Dir("../"))
	mux.Handle("/", http.StripPrefix("/", fs))

	// API endpoints (prefixed to avoid conflicts with static files)
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("POST /api/analyze", handleAnalyze)
	mux.HandleFunc("POST /api/tags", handleTags)
	mux.HandleFunc("POST /api/dress", handleDress)
	mux.HandleFunc("POST /api/remove-bg", handleRemoveBg)
	mux.HandleFunc("POST /api/suggest-outfit", handleSuggestOutfit)
	mux.HandleFunc("POST /api/color-match", handleColorMatch)
	mux.HandleFunc("POST /api/style-match", handleStyleMatch)
	mux.HandleFunc("POST /api/rate-outfit", handleRateOutfit)
	mux.HandleFunc("POST /api/wardrobe-gaps", handleWardrobeGaps)
	mux.HandleFunc("GET /api/status", handleStatus)

	// Database CRUD endpoints
	mux.HandleFunc("GET /api/items", handleListItems)
	mux.HandleFunc("GET /api/items/{id}", handleGetItem)
	mux.HandleFunc("POST /api/items", handleCreateItem)
	mux.HandleFunc("PUT /api/items/{id}", handleUpdateItem)
	mux.HandleFunc("DELETE /api/items/{id}", handleDeleteItem)

	// S3 upload endpoint
	mux.HandleFunc("POST /api/upload", handleS3Upload)

	// Wrap with CORS middleware
	handler := corsMiddleware(mux)

	port := getEnv("PORT", "8556")
	log.Printf("Starting mockup server on :%s", port)
	log.Printf("📂 Serving mockups from: http://localhost:%s", port)
	log.Printf("🤖 Ollama URL: %s", ollamaURL)
	log.Printf("👁️  Vision model: %s", visionModel)
	log.Printf("📝 Text model: %s", textModel)
	log.Printf("🖼️  Rembg URL: %s", rembgURL)
	log.Printf("")
	log.Printf("Open http://localhost:%s/upload.html to test AI features", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	// Check if Ollama is running and what models are available
	resp, err := http.Get(ollamaURL + "/api/tags")
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "Ollama not reachable: "+err.Error())
		return
	}
	defer resp.Body.Close()

	var tagsResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to parse Ollama response")
		return
	}

	models := make([]string, len(tagsResp.Models))
	for i, m := range tagsResp.Models {
		models[i] = m.Name
	}

	hasVision := contains(models, visionModel)
	hasText := contains(models, textModel)

	// Check if rembg is available
	rembgAvailable := false
	if resp, err := http.Get(rembgURL); err == nil {
		resp.Body.Close()
		rembgAvailable = resp.StatusCode == http.StatusOK
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ollama_url":         ollamaURL,
		"vision_model":       visionModel,
		"text_model":         textModel,
		"rembg_url":          rembgURL,
		"available_models":   models,
		"vision_model_ready": hasVision,
		"text_model_ready":   hasText,
		"rembg_available":    rembgAvailable,
		"ready":              hasVision,
	})
}

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Image == "" {
		writeError(w, http.StatusBadRequest, "Image is required")
		return
	}

	// Strip data URI prefix if present
	imageData := req.Image
	if idx := strings.Index(imageData, ","); idx != -1 {
		imageData = imageData[idx+1:]
	}

	// Validate base64
	if _, err := base64.StdEncoding.DecodeString(imageData); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid base64 image data")
		return
	}

	prompt := req.Prompt
	if prompt == "" {
		prompt = `Look at this clothing item. Tell me:

1. What type is it? (Tops, Bottoms, Outerwear, Shoes, or Accessories)
2. What color is it?
3. Read any text on labels or tags - what brand name do you see?
4. Describe it in one sentence
5. List 3-5 style tags

Respond ONLY with JSON:
{"description": "A beige cable-knit sweater", "category": "Tops", "color": "Beige", "brand": "MR MARVIS", "tags": ["casual", "knitwear", "warm"]}

If no brand text visible, use "brand": ""`
	}

	// Allow model override
	modelToUse := visionModel
	if req.Model != "" {
		modelToUse = req.Model
		log.Printf("Using custom model: %s (instead of default %s)", modelToUse, visionModel)
	}

	log.Printf("Analyzing image with %s...", modelToUse)
	log.Printf("Using prompt: %s", prompt)
	start := time.Now()

	ollamaReq := OllamaRequest{
		Model:  modelToUse,
		Prompt: prompt,
		Images: []string{imageData},
		Stream: false,
	}

	reqBody, _ := json.Marshal(ollamaReq)
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		log.Printf("ERROR: Ollama request failed: %v", err)
		writeError(w, http.StatusServiceUnavailable, "Ollama request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("ERROR: Ollama returned status %d: %s", resp.StatusCode, string(body))
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("Ollama error (status %d): %s", resp.StatusCode, string(body)))
		return
	}

	body, _ := io.ReadAll(resp.Body)
	log.Printf("Ollama response time: %v", time.Since(start))
	log.Printf("Raw Ollama response: %s", string(body))

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		log.Printf("ERROR: Failed to parse Ollama JSON: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to parse Ollama response: "+err.Error())
		return
	}

	log.Printf("Parsed response: %s", ollamaResp.Response)

	// Try to parse the response as JSON
	response := parseClothingResponse(ollamaResp.Response)
	response.RawResponse = ollamaResp.Response

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleTags(w http.ResponseWriter, r *http.Request) {
	var req TagsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	if req.Description == "" {
		writeError(w, http.StatusBadRequest, "Description is required")
		return
	}

	prompt := fmt.Sprintf(`Given this clothing item description: "%s"

Generate 5 relevant tags for this item. Tags should be single words or short phrases like: casual, formal, summer, winter, cotton, slim-fit, vintage, etc.

Respond with ONLY a JSON array of strings, nothing else. Example: ["casual", "cotton", "summer"]`, req.Description)

	ollamaReq := OllamaRequest{
		Model:  textModel,
		Prompt: prompt,
		Stream: false,
	}

	reqBody, _ := json.Marshal(ollamaReq)
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "Ollama request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to parse Ollama response")
		return
	}

	// Parse tags from response
	tags := parseTagsResponse(ollamaResp.Response)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagsResponse{Tags: tags})
}

func handleDress(w http.ResponseWriter, r *http.Request) {
	var req DressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.BodyImage == "" {
		writeError(w, http.StatusBadRequest, "Body image is required")
		return
	}

	if len(req.ClothingItems) == 0 {
		writeError(w, http.StatusBadRequest, "At least one clothing item is required")
		return
	}

	// Strip data URI prefixes
	bodyData := req.BodyImage
	if idx := strings.Index(bodyData, ","); idx != -1 {
		bodyData = bodyData[idx+1:]
	}

	prompt := req.Prompt
	if prompt == "" {
		itemsList := ""
		for i, item := range req.ClothingItems {
			itemsList += fmt.Sprintf("%d. %s (%s)\n", i+1, item.Name, item.Category)
		}

		prompt = fmt.Sprintf(`Look at this person/mannequin image and the following clothing items:
%s

For each item, suggest where it should be positioned on the body:
- X: horizontal position (0-100, where 0=left, 50=center, 100=right)
- Y: vertical position (0-100, where 0=top, 50=middle, 100=bottom)
- Z: layer order (higher numbers are on top)

Consider:
- Tops should be positioned on the upper body
- Bottoms on the lower body
- Outerwear should be layered on top
- Proper positioning based on body proportions

Respond ONLY with valid JSON:
{"positions": [{"id": 1, "x": 45, "y": 30, "z": 1}, ...], "explanation": "Brief explanation"}`, itemsList)
	}

	log.Printf("AI Dress request with %d items", len(req.ClothingItems))
	start := time.Now()

	ollamaReq := OllamaRequest{
		Model:  visionModel,
		Prompt: prompt,
		Images: []string{bodyData},
		Stream: false,
	}

	reqBody, _ := json.Marshal(ollamaReq)
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		log.Printf("ERROR: Ollama request failed: %v", err)
		writeError(w, http.StatusServiceUnavailable, "Ollama request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("ERROR: Ollama returned status %d: %s", resp.StatusCode, string(body))
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("Ollama error (status %d)", resp.StatusCode))
		return
	}

	body, _ := io.ReadAll(resp.Body)
	log.Printf("AI Dress response time: %v", time.Since(start))

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		log.Printf("ERROR: Failed to parse Ollama JSON: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to parse Ollama response")
		return
	}

	// Parse the response
	dressResp := parseDressResponse(ollamaResp.Response, len(req.ClothingItems))
	dressResp.RawResponse = ollamaResp.Response

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dressResp)
}

func parseClothingResponse(response string) AnalyzeResponse {
	result := AnalyzeResponse{
		Description: "Unable to analyze",
		Category:    "Unknown",
		Color:       "Unknown",
		Brand:       "",
		Tags:        []string{},
	}

	// Try to find JSON in the response
	response = strings.TrimSpace(response)

	// Remove markdown code fences if present
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Find JSON object boundaries
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start == -1 || end == -1 || end <= start {
		// No JSON found, use raw response as description
		result.Description = response
		return result
	}

	jsonStr := response[start : end+1]

	var parsed struct {
		Description string   `json:"description"`
		Category    string   `json:"category"`
		Color       interface{} `json:"color"` // Can be string or array
		Brand       string   `json:"brand"`
		Tags        []string `json:"tags"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
		if parsed.Description != "" {
			result.Description = parsed.Description
		}
		if parsed.Category != "" {
			result.Category = parsed.Category
		}

		// Handle color as string or array
		switch v := parsed.Color.(type) {
		case string:
			if v != "" {
				result.Color = v
			}
		case []interface{}:
			if len(v) > 0 {
				if str, ok := v[0].(string); ok {
					result.Color = str
				}
			}
		}

		if parsed.Brand != "" {
			result.Brand = parsed.Brand
		}
		if len(parsed.Tags) > 0 {
			result.Tags = parsed.Tags
		}
	}

	return result
}

func parseTagsResponse(response string) []string {
	response = strings.TrimSpace(response)

	// Find JSON array boundaries
	start := strings.Index(response, "[")
	end := strings.LastIndex(response, "]")

	if start == -1 || end == -1 || end <= start {
		return []string{}
	}

	jsonStr := response[start : end+1]

	var tags []string
	if err := json.Unmarshal([]byte(jsonStr), &tags); err == nil {
		return tags
	}

	return []string{}
}

func parseDressResponse(response string, expectedCount int) DressResponse {
	result := DressResponse{
		Positions:   []ItemPosition{},
		Explanation: "",
	}

	// Remove markdown code fences
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	// Find JSON object boundaries
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start == -1 || end == -1 || end <= start {
		// Fallback: create default positions
		for i := 0; i < expectedCount; i++ {
			result.Positions = append(result.Positions, ItemPosition{
				ID: i + 1,
				X:  45 + float64(i*5),
				Y:  30 + float64(i*15),
				Z:  i,
			})
		}
		result.Explanation = "Using default positions (AI response was not JSON)"
		return result
	}

	jsonStr := response[start : end+1]

	var parsed struct {
		Positions   []ItemPosition `json:"positions"`
		Explanation string         `json:"explanation"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
		result.Positions = parsed.Positions
		result.Explanation = parsed.Explanation
	} else {
		// Fallback
		for i := 0; i < expectedCount; i++ {
			result.Positions = append(result.Positions, ItemPosition{
				ID: i + 1,
				X:  45 + float64(i*5),
				Y:  30 + float64(i*15),
				Z:  i,
			})
		}
		result.Explanation = "Using default positions (failed to parse JSON)"
	}

	return result
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func handleRemoveBg(w http.ResponseWriter, r *http.Request) {
	var req RemoveBgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.Image == "" {
		writeError(w, http.StatusBadRequest, "Image is required")
		return
	}

	// Strip data URI prefix if present
	imageData := req.Image
	if idx := strings.Index(imageData, ","); idx != -1 {
		imageData = imageData[idx+1:]
	}

	// Decode base64 to raw bytes
	imgBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid base64 image data")
		return
	}

	log.Printf("Removing background from image (%d bytes)...", len(imgBytes))
	start := time.Now()

	// Call rembg service with multipart/form-data
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "image.png")
	if err != nil {
		log.Printf("ERROR: Failed to create form file: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to prepare request")
		return
	}

	if _, err := part.Write(imgBytes); err != nil {
		log.Printf("ERROR: Failed to write image data: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to prepare request")
		return
	}

	if err := writer.Close(); err != nil {
		log.Printf("ERROR: Failed to close writer: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to prepare request")
		return
	}

	resp, err := http.Post(rembgURL+"/api/remove", writer.FormDataContentType(), &buf)
	if err != nil {
		log.Printf("ERROR: rembg request failed: %v", err)
		writeError(w, http.StatusServiceUnavailable, "Background removal service not available: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("ERROR: rembg returned status %d: %s", resp.StatusCode, string(body))
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("Background removal failed (status %d)", resp.StatusCode))
		return
	}

	// Read the processed image
	processedBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("ERROR: Failed to read rembg response: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to read processed image")
		return
	}

	log.Printf("Background removed in %v (%d bytes -> %d bytes)", time.Since(start), len(imgBytes), len(processedBytes))

	// Encode back to base64
	processedBase64 := base64.StdEncoding.EncodeToString(processedBytes)

	// Return with data URI prefix for PNG (rembg returns PNG with transparency)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RemoveBgResponse{
		Image: "data:image/png;base64," + processedBase64,
	})
}

func handleSuggestOutfit(w http.ResponseWriter, r *http.Request) {
	var req SuggestOutfitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "At least one item required")
		return
	}

	// Build item list for AI with full details
	itemsList := ""
	for _, item := range req.Items {
		tags := ""
		if len(item.Tags) > 0 {
			tags = " [" + strings.Join(item.Tags, ", ") + "]"
		}

		brand := ""
		if item.Brand != "" {
			brand = " by " + item.Brand
		}

		desc := ""
		if item.Description != "" {
			desc = " - " + item.Description
		}

		worn := ""
		if item.WearCount > 0 {
			worn = fmt.Sprintf(" (worn %dx)", item.WearCount)
		}

		itemsList += fmt.Sprintf("- ID %d: %s%s (%s, %s)%s%s%s\n",
			item.ID, item.Name, brand, item.Category, item.Color, tags, desc, worn)
	}

	prompt := fmt.Sprintf(`Help me pick an outfit from my wardrobe.

Occasion: %s
Weather: %s

Available items (%d):
%s

Select items for a complete outfit. Include:
- ONE top (shirt, t-shirt, blouse)
- ONE bottom (pants, jeans, skirt, shorts)
- ONE pair of shoes
- OPTIONAL: One outerwear (jacket, coat, cardigan, sweater) - ONLY if weather requires it
- OPTIONAL: 1-2 accessories (watch, belt, scarf, hat, bag)

Important:
- Choose colors that complement each other
- Consider weather (add layers if cold, lighter items if warm)
- Match formality to occasion
- Total: 3-6 items maximum

Respond with JSON:
{
  "selected_items": [1, 5, 7],
  "reasoning": "Why these work together (mention colors and weather)",
  "tips": ["Styling tip 1", "Styling tip 2"]
}`, req.Occasion, req.Weather, len(req.Items), itemsList)

	log.Printf("Suggesting outfit for %s / %s with %d items", req.Occasion, req.Weather, len(req.Items))
	start := time.Now()

	ollamaReq := OllamaRequest{
		Model:  textModel,
		Prompt: prompt,
		Stream: false,
	}

	reqBody, _ := json.Marshal(ollamaReq)
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		log.Printf("ERROR: Ollama request failed: %v", err)
		writeError(w, http.StatusServiceUnavailable, "AI request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		log.Printf("ERROR: Failed to parse Ollama response: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to parse AI response")
		return
	}

	log.Printf("Outfit suggestion time: %v", time.Since(start))

	result := parseSuggestOutfitResponse(ollamaResp.Response)
	result.RawResponse = ollamaResp.Response

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleColorMatch(w http.ResponseWriter, r *http.Request) {
	var req ColorMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if len(req.Items) < 2 {
		writeError(w, http.StatusBadRequest, "At least 2 items required for color matching")
		return
	}

	itemsList := ""
	for i, item := range req.Items {
		itemsList += fmt.Sprintf("%d. %s: %s color\n", i+1, item.Name, item.Color)
	}

	prompt := fmt.Sprintf(`Analyze the color coordination of these clothing items:

%s

Rate the color compatibility from 0-100 and explain:
- Do these colors work well together?
- Any color clashes?
- Suggestions to improve the color scheme?

Respond ONLY with JSON:
{
  "compatibility_score": 85,
  "analysis": "Brief analysis of color coordination",
  "suggestions": ["Suggestion 1", "Suggestion 2"]
}`, itemsList)

	log.Printf("Analyzing color match for %d items", len(req.Items))
	start := time.Now()

	ollamaReq := OllamaRequest{
		Model:  textModel,
		Prompt: prompt,
		Stream: false,
	}

	reqBody, _ := json.Marshal(ollamaReq)
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "AI request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to parse AI response")
		return
	}

	log.Printf("Color match analysis time: %v", time.Since(start))

	result := parseColorMatchResponse(ollamaResp.Response)
	result.RawResponse = ollamaResp.Response

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleStyleMatch(w http.ResponseWriter, r *http.Request) {
	var req StyleMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if req.BaseItem.ID == 0 {
		writeError(w, http.StatusBadRequest, "Base item required")
		return
	}

	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "Items to match required")
		return
	}

	maxItems := req.MaxItems
	if maxItems == 0 {
		maxItems = 5
	}

	baseTags := strings.Join(req.BaseItem.Tags, ", ")
	itemsList := ""
	for _, item := range req.Items {
		tags := strings.Join(item.Tags, ", ")
		itemsList += fmt.Sprintf("- ID %d: %s (%s, %s) [%s]\n", item.ID, item.Name, item.Category, item.Color, tags)
	}

	prompt := fmt.Sprintf(`I have this clothing item:
%s (%s, %s) [%s]

Find the top %d items from this wardrobe that would match well:
%s

Consider:
- Color coordination
- Style compatibility (tags/aesthetics)
- Formality level
- Seasonal appropriateness

For each matching item, give a score (0-100) and explain why it matches.

Respond ONLY with JSON:
{
  "matching_items": [
    {"id": 5, "name": "Item name", "reason": "Why it matches", "score": 85},
    {"id": 7, "name": "Item name", "reason": "Why it matches", "score": 78}
  ]
}`, req.BaseItem.Name, req.BaseItem.Category, req.BaseItem.Color, baseTags, maxItems, itemsList)

	log.Printf("Finding style matches for item %d", req.BaseItem.ID)
	start := time.Now()

	ollamaReq := OllamaRequest{
		Model:  textModel,
		Prompt: prompt,
		Stream: false,
	}

	reqBody, _ := json.Marshal(ollamaReq)
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "AI request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to parse AI response")
		return
	}

	log.Printf("Style match time: %v", time.Since(start))

	result := parseStyleMatchResponse(ollamaResp.Response)
	result.RawResponse = ollamaResp.Response

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleRateOutfit(w http.ResponseWriter, r *http.Request) {
	var req RateOutfitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "At least one item required")
		return
	}

	itemsList := ""
	for _, item := range req.Items {
		tags := strings.Join(item.Tags, ", ")
		itemsList += fmt.Sprintf("- %s (%s, %s) [%s]\n", item.Name, item.Category, item.Color, tags)
	}

	prompt := fmt.Sprintf(`Rate this outfit from 0-10 and provide constructive feedback:

%s

Consider:
- Color harmony
- Style cohesion
- Appropriate layering
- Balance and proportions

Respond ONLY with JSON:
{
  "rating": 8,
  "feedback": "Overall assessment in 1-2 sentences",
  "strengths": ["What works well", "Another strength"],
  "improvements": ["How to improve", "Another suggestion"]
}`, itemsList)

	log.Printf("Rating outfit with %d items", len(req.Items))
	start := time.Now()

	ollamaReq := OllamaRequest{
		Model:  textModel,
		Prompt: prompt,
		Stream: false,
	}

	reqBody, _ := json.Marshal(ollamaReq)
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "AI request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to parse AI response")
		return
	}

	log.Printf("Outfit rating time: %v", time.Since(start))

	result := parseRateOutfitResponse(ollamaResp.Response)
	result.RawResponse = ollamaResp.Response

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func handleWardrobeGaps(w http.ResponseWriter, r *http.Request) {
	var req WardrobeGapsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	if len(req.Items) == 0 {
		writeError(w, http.StatusBadRequest, "At least one item required")
		return
	}

	// Group items by category
	categoryCount := make(map[string]int)
	colorCount := make(map[string]int)
	itemsList := ""

	for _, item := range req.Items {
		categoryCount[item.Category]++
		colorCount[item.Color]++
		tags := strings.Join(item.Tags, ", ")
		itemsList += fmt.Sprintf("- %s (%s, %s) [%s]\n", item.Name, item.Category, item.Color, tags)
	}

	categorySummary := ""
	for cat, count := range categoryCount {
		categorySummary += fmt.Sprintf("- %s: %d items\n", cat, count)
	}

	prompt := fmt.Sprintf(`Analyze this wardrobe and identify gaps or missing essentials:

Total items: %d

By category:
%s

All items:
%s

Provide:
1. Overall assessment of wardrobe versatility
2. Missing essential items that would increase outfit options
3. What's well-covered
4. General wardrobe building tips

Respond ONLY with JSON:
{
  "summary": "Overall wardrobe assessment in 2-3 sentences",
  "missing_items": ["Essential item 1", "Essential item 2", "Versatile piece 3"],
  "strengths": ["What's well covered", "Another strength"],
  "tips": ["Wardrobe tip 1", "Wardrobe tip 2"]
}`, len(req.Items), categorySummary, itemsList)

	log.Printf("Analyzing wardrobe gaps for %d items", len(req.Items))
	start := time.Now()

	ollamaReq := OllamaRequest{
		Model:  textModel,
		Prompt: prompt,
		Stream: false,
	}

	reqBody, _ := json.Marshal(ollamaReq)
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "AI request failed: "+err.Error())
		return
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to parse AI response")
		return
	}

	log.Printf("Wardrobe gaps analysis time: %v", time.Since(start))

	result := parseWardrobeGapsResponse(ollamaResp.Response)
	result.RawResponse = ollamaResp.Response

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Parsing functions for AI responses

func parseSuggestOutfitResponse(response string) SuggestOutfitResponse {
	result := SuggestOutfitResponse{
		SelectedItems: []int{},
		Reasoning:     "No outfit suggested",
		Tips:          []string{},
	}

	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start == -1 || end == -1 {
		return result
	}

	jsonStr := response[start : end+1]

	var parsed struct {
		SelectedItems []int    `json:"selected_items"`
		Reasoning     string   `json:"reasoning"`
		Tips          []string `json:"tips"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
		result.SelectedItems = parsed.SelectedItems
		if parsed.Reasoning != "" {
			result.Reasoning = parsed.Reasoning
		}
		if len(parsed.Tips) > 0 {
			result.Tips = parsed.Tips
		}
	}

	return result
}

func parseColorMatchResponse(response string) ColorMatchResponse {
	result := ColorMatchResponse{
		CompatibilityScore: 50,
		Analysis:           "Unable to analyze colors",
		Suggestions:        []string{},
	}

	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start == -1 || end == -1 {
		return result
	}

	jsonStr := response[start : end+1]

	var parsed struct {
		CompatibilityScore int      `json:"compatibility_score"`
		Analysis           string   `json:"analysis"`
		Suggestions        []string `json:"suggestions"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
		if parsed.CompatibilityScore > 0 {
			result.CompatibilityScore = parsed.CompatibilityScore
		}
		if parsed.Analysis != "" {
			result.Analysis = parsed.Analysis
		}
		if len(parsed.Suggestions) > 0 {
			result.Suggestions = parsed.Suggestions
		}
	}

	return result
}

func parseStyleMatchResponse(response string) StyleMatchResponse {
	result := StyleMatchResponse{
		MatchingItems: []MatchedItem{},
	}

	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start == -1 || end == -1 {
		return result
	}

	jsonStr := response[start : end+1]

	var parsed struct {
		MatchingItems []MatchedItem `json:"matching_items"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
		result.MatchingItems = parsed.MatchingItems
	}

	return result
}

func parseRateOutfitResponse(response string) RateOutfitResponse {
	result := RateOutfitResponse{
		Rating:       5,
		Feedback:     "Unable to rate outfit",
		Strengths:    []string{},
		Improvements: []string{},
	}

	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start == -1 || end == -1 {
		return result
	}

	jsonStr := response[start : end+1]

	var parsed struct {
		Rating       int      `json:"rating"`
		Feedback     string   `json:"feedback"`
		Strengths    []string `json:"strengths"`
		Improvements []string `json:"improvements"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
		if parsed.Rating > 0 {
			result.Rating = parsed.Rating
		}
		if parsed.Feedback != "" {
			result.Feedback = parsed.Feedback
		}
		if len(parsed.Strengths) > 0 {
			result.Strengths = parsed.Strengths
		}
		if len(parsed.Improvements) > 0 {
			result.Improvements = parsed.Improvements
		}
	}

	return result
}

func parseWardrobeGapsResponse(response string) WardrobeGapsResponse {
	result := WardrobeGapsResponse{
		Summary:      "Unable to analyze wardrobe",
		MissingItems: []string{},
		Strengths:    []string{},
		Tips:         []string{},
	}

	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start == -1 || end == -1 {
		return result
	}

	jsonStr := response[start : end+1]

	var parsed struct {
		Summary      string   `json:"summary"`
		MissingItems []string `json:"missing_items"`
		Strengths    []string `json:"strengths"`
		Tips         []string `json:"tips"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
		if parsed.Summary != "" {
			result.Summary = parsed.Summary
		}
		if len(parsed.MissingItems) > 0 {
			result.MissingItems = parsed.MissingItems
		}
		if len(parsed.Strengths) > 0 {
			result.Strengths = parsed.Strengths
		}
		if len(parsed.Tips) > 0 {
			result.Tips = parsed.Tips
		}
	}

	return result
}
