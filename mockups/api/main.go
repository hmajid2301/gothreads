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
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultOllamaURL   = "http://localhost:11434"
	defaultVisionModel = "llava:7b"
	defaultTextModel   = "llama3.2:3b"
)

var (
	ollamaURL   = getEnv("OLLAMA_URL", defaultOllamaURL)
	visionModel = getEnv("OLLAMA_VISION_MODEL", defaultVisionModel)
	textModel   = getEnv("OLLAMA_TEXT_MODEL", defaultTextModel)
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
	mux := http.NewServeMux()

	// Serve static mockup files from parent directory
	fs := http.FileServer(http.Dir("../"))
	mux.Handle("/", http.StripPrefix("/", fs))

	// API endpoints (prefixed to avoid conflicts with static files)
	mux.HandleFunc("GET /api/health", handleHealth)
	mux.HandleFunc("POST /api/analyze", handleAnalyze)
	mux.HandleFunc("POST /api/tags", handleTags)
	mux.HandleFunc("POST /api/dress", handleDress)
	mux.HandleFunc("GET /api/status", handleStatus)

	// Wrap with CORS middleware
	handler := corsMiddleware(mux)

	port := getEnv("PORT", "8556")
	log.Printf("Starting mockup server on :%s", port)
	log.Printf("📂 Serving mockups from: http://localhost:%s", port)
	log.Printf("🤖 Ollama URL: %s", ollamaURL)
	log.Printf("👁️  Vision model: %s", visionModel)
	log.Printf("📝 Text model: %s", textModel)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ollama_url":         ollamaURL,
		"vision_model":       visionModel,
		"text_model":         textModel,
		"available_models":   models,
		"vision_model_ready": hasVision,
		"text_model_ready":   hasText,
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
