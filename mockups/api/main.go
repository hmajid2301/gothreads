// Throwaway PoC - Ollama API proxy for mockups
// Usage: go run main.go
// Requires: Ollama running on localhost:11434 with llava model

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-pkgz/routegroup"
	"github.com/gofrs/uuid/v5"
	"github.com/joho/godotenv"
)

const (
	defaultOllamaURL         = "http://localhost:11434"
	defaultVisionModel       = "llava:34b"
	defaultTextModel         = "llama3.2:3b"
	defaultSpatialModelLocal = "qwen3-vl:32b"
	defaultSpatialModelCloud = "qwen3-vl:235b-cloud"
	defaultTextModelCloud    = "deepseek-v3.1:671b-cloud"
	defaultRembgURL          = "http://localhost:5000"
	// defaultVtonBackendURL removed - using local processing only
	defaultCatvtonLocalURL  = "http://localhost:7860"
	defaultLadivtonShoesURL = "http://localhost:8558"
)

var (
	ollamaURL        string
	ollamaCloud      bool
	visionModel      string
	textModel        string
	cloudTextModel   string
	spatialModel     string
	rembgURL         string
	vtonBackendURL   string
	catvtonLocalURL  string
	ladivtonShoesURL string
)

func init() {
	// Load .env before reading env vars
	_ = godotenv.Load()

	// Load config from YAML (with env overrides)
	cfg, err := loadConfig()
	if err != nil {
		log.Printf("⚠️  Failed to load config: %v, using defaults", err)
		cfg = getDefaultConfig()
	}
	appConfig = cfg

	// Set global variables from config for backward compatibility
	ollamaURL = cfg.AI.Ollama.URL
	visionModel = cfg.AI.Ollama.VisionModel
	textModel = cfg.AI.Ollama.TextModel
	spatialModel = cfg.AI.Ollama.SpatialModel
	rembgURL = cfg.ImageProcessing.RembgURL
	vtonBackendURL = cfg.ImageProcessing.VTONBackend
	catvtonLocalURL = cfg.ImageProcessing.CatVTONURL
	ladivtonShoesURL = cfg.ImageProcessing.LadiVTONShoes

	// Detect cloud models by checking if they're pulled in Ollama
	ollamaCloud = isModelAvailable(defaultSpatialModelCloud) || isModelAvailable(defaultTextModelCloud)

	if ollamaCloud {
		cloudTextModel = defaultTextModelCloud
	} else {
		cloudTextModel = textModel
	}
}

// isModelAvailable checks if a model is pulled in Ollama
func isModelAvailable(model string) bool {
	resp, err := http.Post(ollamaURL+"/api/show", "application/json",
		bytes.NewReader([]byte(fmt.Sprintf(`{"name":"%s"}`, model))))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

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
	System string   `json:"system,omitempty"`
	Images []string `json:"images,omitempty"` // base64 encoded images
	Stream bool     `json:"stream"`
}

// OllamaResponse is the response format from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// ============================================================================
// OLLAMA CHAT API TYPES (for tool calling via /api/chat)
// ============================================================================

// OllamaChatMessage represents a single message in the chat conversation
type OllamaChatMessage struct {
	Role      string           `json:"role"` // system, user, assistant, tool
	Content   string           `json:"content"`
	ToolCalls []OllamaToolCall `json:"tool_calls,omitempty"` // only in assistant messages
}

// OllamaToolCall represents a tool call from the AI
type OllamaToolCall struct {
	Function OllamaFunctionCall `json:"function"`
}

// OllamaFunctionCall contains the function name and arguments
type OllamaFunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// OllamaToolDef defines a tool the AI can call
type OllamaToolDef struct {
	Type     string            `json:"type"` // "function"
	Function OllamaFunctionDef `json:"function"`
}

// OllamaFunctionDef describes a function's name, description, and parameters
type OllamaFunctionDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// OllamaChatRequest is the request format for /api/chat
type OllamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []OllamaChatMessage `json:"messages"`
	Tools    []OllamaToolDef     `json:"tools,omitempty"`
	Stream   bool                `json:"stream"`
}

// OllamaChatResponse is the response from /api/chat
type OllamaChatResponse struct {
	Message OllamaChatMessage `json:"message"`
	Done    bool              `json:"done"`
}

// ChatRequest for AI chat with tool calling
type ChatRequest struct {
	Message string              `json:"message"`           // user's message
	History []OllamaChatMessage `json:"history,omitempty"` // previous conversation
}

// ChatResponse returned to frontend
type ChatResponse struct {
	Message   string `json:"message"`    // AI's text response
	ToolsUsed int    `json:"tools_used"` // how many tool calls were made
}

// AnalyzeRequest is what the frontend sends
type AnalyzeRequest struct {
	Image  string `json:"image"`  // base64 encoded image (with or without data URI prefix)
	Prompt string `json:"prompt"` // optional custom prompt
	Model  string `json:"model"`  // optional model override (e.g., "llava:13b")
}

// AnalyzeResponse is what we send back to frontend
type AnalyzeResponse struct {
	Name        string   `json:"name"`
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

// TryOnRequest for virtual try-on
type TryOnRequest struct {
	PersonImage  string      `json:"person_image"`  // base64 encoded person image
	GarmentImage string      `json:"garment_image"` // base64 encoded garment image (legacy/single)
	ClothType    string      `json:"cloth_type"`    // upper, lower, overall (legacy/single)
	Garments     []TryOnItem `json:"garments"`      // List of garments for multi-try-on
}

type TryOnItem struct {
	Image     string `json:"image"`      // base64 encoded garment image
	ClothType string `json:"cloth_type"` // upper, lower, overall
}

// TryOnResponse returns the try-on result
type TryOnResponse struct {
	Image string `json:"image"` // base64 encoded result image
}

// AIJob represents any async AI operation
type AIJob struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`   // analyze, tags, dress, remove-bg, tryon, suggest-outfit, color-match, style-match, rate-outfit, wardrobe-gaps
	Status    string          `json:"status"` // queued, processing, complete, failed, cancelled
	Message   string          `json:"message"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     string          `json:"error,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	cancel    context.CancelFunc
	mu        sync.RWMutex
	listeners []chan AIJobEvent
}

// AIJobEvent is sent to SSE listeners
type AIJobEvent struct {
	Event string // status, result, error
	Data  string
}

// AIJobSubmitRequest is what the frontend sends to create a job
type AIJobSubmitRequest struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

var (
	aiJobs   = make(map[string]*AIJob)
	aiJobsMu sync.RWMutex
)

func generateJobID() string {
	return uuid.Must(uuid.NewV4()).String()
}

func (j *AIJob) updateStatus(status, message string) {
	j.mu.Lock()
	j.Status = status
	j.Message = message
	listeners := make([]chan AIJobEvent, len(j.listeners))
	copy(listeners, j.listeners)
	j.mu.Unlock()

	log.Printf("🤖 Job %s [%s]: [%s] %s", j.ID[:8], j.Type, status, message)

	for _, ch := range listeners {
		select {
		case ch <- AIJobEvent{Event: "status", Data: message}:
		default:
		}
	}
}

func (j *AIJob) complete(result interface{}) {
	resultJSON, _ := json.Marshal(result)
	j.mu.Lock()
	j.Status = "complete"
	j.Result = resultJSON
	j.Message = "Complete"
	listeners := make([]chan AIJobEvent, len(j.listeners))
	copy(listeners, j.listeners)
	j.mu.Unlock()

	log.Printf("🤖 Job %s [%s]: complete", j.ID[:8], j.Type)

	for _, ch := range listeners {
		select {
		case ch <- AIJobEvent{Event: "result", Data: string(resultJSON)}:
		default:
		}
	}
}

func (j *AIJob) fail(errMsg string) {
	j.mu.Lock()
	j.Status = "failed"
	j.Error = errMsg
	j.Message = errMsg
	listeners := make([]chan AIJobEvent, len(j.listeners))
	copy(listeners, j.listeners)
	j.mu.Unlock()

	log.Printf("🤖 Job %s [%s]: failed - %s", j.ID[:8], j.Type, errMsg)

	for _, ch := range listeners {
		select {
		case ch <- AIJobEvent{Event: "error", Data: errMsg}:
		default:
		}
	}
}

func (j *AIJob) addListener() chan AIJobEvent {
	ch := make(chan AIJobEvent, 10)
	j.mu.Lock()
	j.listeners = append(j.listeners, ch)
	j.mu.Unlock()
	return ch
}

func (j *AIJob) removeListener(ch chan AIJobEvent) {
	j.mu.Lock()
	for i, l := range j.listeners {
		if l == ch {
			j.listeners = append(j.listeners[:i], j.listeners[i+1:]...)
			break
		}
	}
	j.mu.Unlock()
	close(ch)
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
	BaseItem ClothingItemForMatch   `json:"base_item"` // item to match with
	Items    []ClothingItemForMatch `json:"items"`     // available items
	MaxItems int                    `json:"max_items"` // max items to return (default 5)
}

// StyleMatchResponse returns matching items
type StyleMatchResponse struct {
	MatchingItems []MatchedItem `json:"matching_items"`
	RawResponse   string        `json:"raw_response,omitempty"`
}

type MatchedItem struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Reason string `json:"reason"` // Why it matches
	Score  int    `json:"score"`  // 0-100 match score
}

// RateOutfitRequest for outfit feedback
type RateOutfitRequest struct {
	Items []ClothingItemForMatch `json:"items"` // items in outfit
}

// RateOutfitResponse returns outfit rating and feedback
type RateOutfitResponse struct {
	Rating        int      `json:"rating"`         // 0-10
	Feedback      string   `json:"feedback"`       // Overall feedback
	Strengths     []string `json:"strengths"`      // What works well
	Improvements  []string `json:"improvements"`   // What could be better
	ColorScore    int      `json:"color_score"`    // 0-100 color harmony score
	ColorAnalysis string   `json:"color_analysis"` // Color coordination feedback
	RawResponse   string   `json:"raw_response,omitempty"`
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

// DetectOutfitItemsRequest for detecting multiple clothing items in one photo
type DetectOutfitItemsRequest struct {
	Image string `json:"image"` // base64 encoded image
}

// DetectOutfitItemsResponse returns detected clothing items
type DetectOutfitItemsResponse struct {
	Items       []DetectedItem `json:"items"`
	Count       int            `json:"count"`
	Description string         `json:"description"`
	RawResponse string         `json:"raw_response,omitempty"`
}

type DetectedItem struct {
	Category    string `json:"category"`    // tops, bottoms, shoes, accessories, outerwear
	Description string `json:"description"` // e.g., "blue denim jacket", "white t-shirt"
	Color       string `json:"color"`
	Confidence  string `json:"confidence"` // high, medium, low
}

// ScrapeRequest for scraping product info from URL
type ScrapeRequest struct {
	URL string `json:"url"`
}

// ScrapeResponse returns scraped product info
type ScrapeResponse struct {
	Title       string `json:"title"`
	ImageURL    string `json:"image_url"`
	Description string `json:"description"`
	SiteName    string `json:"site_name"`
	OriginalURL string `json:"original_url"`
	Category    string `json:"category,omitempty"`
	Color       string `json:"color,omitempty"`
	Brand       string `json:"brand,omitempty"`
	Price       string `json:"price,omitempty"`
}

func main() {
	// Initialize auth
	initAuth()

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
	router := routegroup.New(mux)

	// Serve static mockup files from parent directory
	fs := http.FileServer(http.Dir("../"))
	router.Handle("/", http.StripPrefix("/", fs))

	// Public endpoints
	router.HandleFunc("GET /api/health", handleHealth)
	router.HandleFunc("GET /api/status", handleStatus)
	router.HandleFunc("GET /api/config", handleGetConfig)
	router.HandleFunc("PUT /api/config", handleUpdateConfig)

	// Auth endpoints (public)
	router.HandleFunc("GET /api/auth/{provider}", handleOAuthLogin)
	router.HandleFunc("GET /api/auth/{provider}/callback", handleOAuthCallback)
	router.HandleFunc("GET /api/auth/logout", handleLogout)
	router.HandleFunc("GET /api/auth/me", handleGetCurrentUser)

	// AI job endpoints (public for now, can add auth later)
	router.HandleFunc("POST /api/ai/jobs", handleAIJobSubmit)
	router.HandleFunc("GET /api/ai/jobs/{id}", handleAIJobStatus)
	router.HandleFunc("DELETE /api/ai/jobs/{id}", handleAIJobCancel)
	router.HandleFunc("POST /api/ai/segment-outfit", handleSegmentOutfit)

	// Analytics endpoints (public for PoC)
	router.HandleFunc("GET /api/analytics/wardrobe-stats", handleWardrobeStats)
	router.HandleFunc("GET /api/analytics/most-worn", handleMostWorn)
	router.HandleFunc("GET /api/analytics/least-worn", handleLeastWorn)

	// Protected endpoints - require authentication
	auth := router.Group()
	auth.Use(authMiddlewareHandler)
	auth.HandleFunc("GET /api/items", handleListItems)
	auth.HandleFunc("GET /api/items/{id}", handleGetItem)
	auth.HandleFunc("POST /api/items", handleCreateItem)
	auth.HandleFunc("PUT /api/items/{id}", handleUpdateItem)
	auth.HandleFunc("DELETE /api/items/{id}", handleDeleteItem)
	auth.HandleFunc("GET /api/outfits", handleListOutfits)
	auth.HandleFunc("GET /api/outfits/{id}", handleGetOutfit)
	auth.HandleFunc("POST /api/outfits", handleCreateOutfit)
	auth.HandleFunc("PUT /api/outfits/{id}", handleUpdateOutfit)
	auth.HandleFunc("DELETE /api/outfits/{id}", handleDeleteOutfit)
	auth.HandleFunc("GET /api/calendar", handleListCalendarEvents)
	auth.HandleFunc("POST /api/calendar", handleUpsertCalendarEvent)
	auth.HandleFunc("DELETE /api/calendar/{date}", handleDeleteCalendarEvent)
	auth.HandleFunc("POST /api/wear", handleLogWear)
	auth.HandleFunc("GET /api/wear", handleGetWearHistory)
	auth.HandleFunc("POST /api/scrape", handleScrapeURL)
	auth.HandleFunc("POST /api/upload", handleS3Upload)

	// Admin endpoints - require admin privileges
	admin := router.Group()
	admin.Use(authMiddlewareHandler, adminMiddlewareHandler)
	admin.HandleFunc("GET /api/admin/users", handleAdminGetUsers)
	admin.HandleFunc("GET /api/admin/stats", handleAdminGetStats)
	admin.HandleFunc("POST /api/admin/users/{id}/approve", handleAdminApproveUser)
	admin.HandleFunc("PUT /api/admin/users/{id}/admin", handleAdminSetAdmin)
	admin.HandleFunc("DELETE /api/admin/users/{id}", handleAdminDeleteUser)
	admin.HandleFunc("DELETE /api/admin/users/{id}/reject", handleAdminRejectUser)
	admin.HandleFunc("POST /api/admin/users/approve-all", handleAdminApproveAll)

	// Wrap with CORS middleware
	handler := corsMiddleware(router)

	port := getEnv("PORT", "8556")
	log.Printf("Starting mockup server on :%s", port)
	log.Printf("📂 Serving mockups from: http://localhost:%s", port)
	log.Printf("🤖 Ollama URL: %s", ollamaURL)
	log.Printf("👁️  Vision model: %s", visionModel)
	log.Printf("📝 Text model: %s", textModel)
	if ollamaCloud {
		log.Printf("☁️  Ollama Cloud models detected")
	}
	log.Printf("📐 Spatial model: %s", spatialModel)
	log.Printf("💬 Cloud text model: %s", cloudTextModel)
	log.Printf("🖼️  Rembg URL: %s", rembgURL)
	log.Printf("🧥 FASHN VTON: %s", vtonBackendURL)
	log.Printf("🧥 CatVTON Local fallback: %s", catvtonLocalURL)
	log.Printf("👟 LaDI-VTON Shoes: %s", ladivtonShoesURL)
	log.Printf("")
	log.Printf("Open http://localhost:%s/wardrobe.html to get started", port)

	// Cleanup old jobs every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			cutoff := time.Now().Add(-30 * time.Minute)
			aiJobsMu.Lock()
			for id, job := range aiJobs {
				job.mu.RLock()
				done := (job.Status == "complete" || job.Status == "failed" || job.Status == "cancelled") && job.CreatedAt.Before(cutoff)
				job.mu.RUnlock()
				if done {
					delete(aiJobs, id)
				}
			}
			aiJobsMu.Unlock()
		}
	}()

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

// Middleware handlers for routegroup
func authMiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authConfig.SkipAuth {
			next.ServeHTTP(w, r)
			return
		}

		user := getUserFromRequest(r)
		if user == nil {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				writeError(w, http.StatusUnauthorized, "Authentication required")
				return
			}
			http.Redirect(w, r, "/login.html", http.StatusSeeOther)
			return
		}

		if !user.IsApproved {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				writeError(w, http.StatusForbidden, "Account pending approval")
				return
			}
			http.Redirect(w, r, "/pending.html", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func adminMiddlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromRequest(r)
		if user == nil || !user.IsAdmin {
			writeError(w, http.StatusForbidden, "Admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
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

	hasSpatial := contains(models, spatialModel) || strings.HasSuffix(spatialModel, "-cloud")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"ollama_url":          ollamaURL,
		"vision_model":        visionModel,
		"text_model":          textModel,
		"spatial_model":       spatialModel,
		"cloud_text_model":    cloudTextModel,
		"ollama_cloud":        ollamaCloud,
		"rembg_url":           rembgURL,
		"available_models":    models,
		"vision_model_ready":  hasVision,
		"text_model_ready":    hasText,
		"spatial_model_ready": hasSpatial,
		"rembg_available":     rembgAvailable,
		"ready":               hasVision,
	})
}

// ============================================================================
// GENERIC AI JOB HANDLERS
// ============================================================================

// aiJobWorkers maps job types to their worker functions
var aiJobWorkers = map[string]func(context.Context, *AIJob, json.RawMessage){
	"analyze":             runAnalyzeJob,
	"tags":                runTagsJob,
	"dress":               runDressJob,
	"remove-bg":           runRemoveBgJob,
	"tryon":               runTryOnJob,
	"suggest-outfit":      runSuggestOutfitJob,
	"color-match":         runColorMatchJob,
	"style-match":         runStyleMatchJob,
	"rate-outfit":         runRateOutfitJob,
	"wardrobe-gaps":       runWardrobeGapsJob,
	"chat":                runChatJob,
	"detect-outfit-items": runDetectOutfitItemsJob,
	"scrape":              runScrapeJob,
}

func handleAIJobSubmit(w http.ResponseWriter, r *http.Request) {
	var req AIJobSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	workerFn, ok := aiJobWorkers[req.Type]
	if !ok {
		writeError(w, http.StatusBadRequest, "Unknown job type: "+req.Type)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	job := &AIJob{
		ID:        generateJobID(),
		Type:      req.Type,
		Status:    "queued",
		Message:   "Job queued",
		CreatedAt: time.Now(),
		cancel:    cancel,
	}

	aiJobsMu.Lock()
	aiJobs[job.ID] = job
	aiJobsMu.Unlock()

	log.Printf("🤖 Job %s [%s] queued", job.ID[:8], req.Type)

	go workerFn(ctx, job, req.Payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"id":     job.ID,
		"status": "queued",
	})
}

func handleAIJobStatus(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")

	aiJobsMu.RLock()
	job, ok := aiJobs[jobID]
	aiJobsMu.RUnlock()

	if !ok {
		writeError(w, http.StatusNotFound, "Job not found")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Send current state first
	job.mu.RLock()
	status := job.Status
	message := job.Message
	result := job.Result
	errMsg := job.Error
	job.mu.RUnlock()

	sendSSE(w, "status", message)

	if status == "complete" && result != nil {
		sendSSE(w, "result", string(result))
		return
	}
	if status == "failed" {
		sendSSE(w, "error", errMsg)
		return
	}
	if status == "cancelled" {
		sendSSE(w, "error", "Job was cancelled")
		return
	}

	ch := job.addListener()
	defer job.removeListener(ch)

	for {
		select {
		case evt, ok := <-ch:
			if !ok {
				return
			}
			sendSSE(w, evt.Event, evt.Data)
			if evt.Event == "result" || evt.Event == "error" {
				return
			}
		case <-r.Context().Done():
			return
		}
	}
}

func handleAIJobCancel(w http.ResponseWriter, r *http.Request) {
	jobID := r.PathValue("id")

	aiJobsMu.RLock()
	job, ok := aiJobs[jobID]
	aiJobsMu.RUnlock()

	if !ok {
		writeError(w, http.StatusNotFound, "Job not found")
		return
	}

	job.mu.Lock()
	if job.Status == "queued" || job.Status == "processing" {
		job.Status = "cancelled"
		job.Message = "Job cancelled by user"
		job.cancel()
		log.Printf("🤖 Job %s [%s] cancelled", job.ID[:8], job.Type)
	}
	job.mu.Unlock()

	job.fail("Job cancelled by user")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"id":     job.ID,
		"status": "cancelled",
	})
}

// ============================================================================
// AI JOB WORKERS
// ============================================================================

func runAnalyzeJob(ctx context.Context, job *AIJob, payload json.RawMessage) {
	var req AnalyzeRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		job.fail("Invalid payload: " + err.Error())
		return
	}

	if req.Image == "" {
		job.fail("Image is required")
		return
	}

	imageData := req.Image
	if idx := strings.Index(imageData, ","); idx != -1 {
		imageData = imageData[idx+1:]
	}

	if _, err := base64.StdEncoding.DecodeString(imageData); err != nil {
		job.fail("Invalid base64 image data")
		return
	}

	job.updateStatus("processing", "Analyzing clothing with AI...")

	prompt := req.Prompt
	if prompt == "" {
		prompt = `Look at this clothing item. Tell me:

1. What type is it? (Tops, Bottoms, Outerwear, Shoes, or Accessories)
2. What color is it?
3. Read any text on labels or tags - what brand name do you see?
4. Give a SHORT name (2-5 words, e.g. "Blue Denim Jeans", "White Cotton T-Shirt")
5. Give a DETAILED description (1-3 sentences covering fit, material, style, notable features)
6. List 3-5 style tags

Respond ONLY with JSON:
{"name": "Beige Cable-Knit Sweater", "description": "A warm cable-knit sweater in beige with a crew neck, ribbed cuffs and hem. Made from a soft wool blend, suitable for layering in autumn and winter. Features a relaxed fit with a classic fisherman knit pattern.", "category": "Tops", "color": "Beige", "brand": "MR MARVIS", "tags": ["casual", "knitwear", "warm"]}

If no brand text visible, use "brand": ""`
	}

	primaryModel := visionModel
	if ollamaCloud {
		primaryModel = defaultSpatialModelCloud
	}
	if req.Model != "" {
		primaryModel = req.Model
	}

	if ctx.Err() != nil {
		return
	}

	start := time.Now()
	ollamaResp, usedModel, err := callOllamaWithFallback(prompt, []string{imageData}, primaryModel, visionModel)
	if err != nil {
		job.fail("Ollama request failed: " + err.Error())
		return
	}

	log.Printf("Analyze job %s: response time %v (model: %s)", job.ID[:8], time.Since(start), usedModel)

	response := parseClothingResponse(ollamaResp.Response)
	response.RawResponse = ollamaResp.Response
	job.complete(response)
}

func runTagsJob(ctx context.Context, job *AIJob, payload json.RawMessage) {
	var req TagsRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		job.fail("Invalid payload: " + err.Error())
		return
	}

	if req.Description == "" {
		job.fail("Description is required")
		return
	}

	job.updateStatus("processing", "Generating tags...")

	prompt := fmt.Sprintf(`Given this clothing item description: "%s"

Generate 5 relevant tags for this item. Tags should be single words or short phrases like: casual, formal, summer, winter, cotton, slim-fit, vintage, etc.

Respond with ONLY a JSON array of strings, nothing else. Example: ["casual", "cotton", "summer"]`, req.Description)

	if ctx.Err() != nil {
		return
	}

	ollamaResp, _, err := callOllamaWithFallback(prompt, nil, cloudTextModel, textModel)
	if err != nil {
		job.fail("Ollama request failed: " + err.Error())
		return
	}

	tags := parseTagsResponse(ollamaResp.Response)
	job.complete(TagsResponse{Tags: tags})
}

func runDressJob(ctx context.Context, job *AIJob, payload json.RawMessage) {
	var req DressRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		job.fail("Invalid payload: " + err.Error())
		return
	}

	if req.BodyImage == "" {
		job.fail("Body image is required")
		return
	}

	if len(req.ClothingItems) == 0 {
		job.fail("At least one clothing item is required")
		return
	}

	job.updateStatus("processing", "AI is positioning clothing on body...")

	bodyData := req.BodyImage
	isSVG := strings.HasPrefix(bodyData, "data:image/svg")
	if idx := strings.Index(bodyData, ","); idx != -1 {
		bodyData = bodyData[idx+1:]
	}

	prompt := req.Prompt
	if prompt == "" {
		itemsList := ""
		for _, item := range req.ClothingItems {
			itemsList += fmt.Sprintf("- ID %d: %s (category: %s)\n", item.ID, item.Name, item.Category)
		}

		prompt = fmt.Sprintf(`Look at this person/mannequin image. I want to overlay clothing items on this body.

The image dimensions map to a coordinate system where:
- X goes from 0 (left edge) to 100 (right edge)
- Y goes from 0 (top edge) to 100 (bottom edge)

For reference on a standing person/mannequin:
- Head/neck area: Y around 5-15
- Shoulders: Y around 15-20
- Chest/torso: Y around 20-40
- Waist: Y around 40-45
- Hips: Y around 45-55
- Upper legs: Y around 55-70
- Lower legs: Y around 70-85
- Feet: Y around 85-95
- Center of body: X around 40-50

Items to position:
%s

For each item, give the X,Y coordinates where its CENTER should be placed on the body, and a Z layer order (higher = on top).

Think step by step about where each category belongs:
- Tops (shirts, t-shirts): center chest area (X ~45, Y ~30)
- Bottoms (pants, jeans): center hip/leg area (X ~45, Y ~55)
- Shoes: feet area (X ~45, Y ~90)
- Outerwear (jacket, coat): over the torso, slightly wider (X ~45, Y ~28), Z should be highest
- Accessories: varies - watches near wrist, hats near head, bags to the side

Respond ONLY with valid JSON in English. Do not use any other language:
{"positions": [{"id": 1, "x": 45, "y": 30, "z": 1}], "explanation": "Brief explanation in English"}`, itemsList)
	}

	var images []string
	if !isSVG {
		images = []string{bodyData}
	} else {
		log.Printf("Body image is SVG, using text-only positioning")
	}

	if ctx.Err() != nil {
		return
	}

	start := time.Now()
	ollamaResp, usedModel, err := callOllamaWithFallback(prompt, images, spatialModel, defaultSpatialModelLocal)
	if err != nil {
		job.fail("Ollama request failed: " + err.Error())
		return
	}

	log.Printf("Dress job %s: response time %v (model: %s)", job.ID[:8], time.Since(start), usedModel)

	dressResp := parseDressResponse(ollamaResp.Response, len(req.ClothingItems))
	dressResp.RawResponse = ollamaResp.Response
	job.complete(dressResp)
}

func parseClothingResponse(response string) AnalyzeResponse {
	result := AnalyzeResponse{
		Name:        "",
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
		Name        string      `json:"name"`
		Description string      `json:"description"`
		Category    string      `json:"category"`
		Color       interface{} `json:"color"` // Can be string or array
		Brand       string      `json:"brand"`
		Tags        []string    `json:"tags"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
		if parsed.Name != "" {
			result.Name = parsed.Name
		}
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

func stripDataURIPrefix(s string) string {
	if idx := strings.Index(s, ","); idx != -1 && strings.HasPrefix(s, "data:") {
		return s[idx+1:]
	}
	return s
}

const fashionSystemPrompt = "You are a helpful fashion consultant and wardrobe assistant. Provide concise, practical advice about clothing, style, and outfit coordination."

func callOllamaWithFallback(prompt string, images []string, primaryModel, fallbackModel string) (*OllamaResponse, string, error) {
	resp, err := callOllama(prompt, images, primaryModel)
	if err == nil {
		return resp, primaryModel, nil
	}

	if primaryModel == fallbackModel {
		return nil, primaryModel, err
	}

	log.Printf("⚠️  Ollama %s failed (%v), falling back to %s", primaryModel, err, fallbackModel)
	resp, err = callOllama(prompt, images, fallbackModel)
	if err != nil {
		return nil, fallbackModel, err
	}
	return resp, fallbackModel, nil
}

func callOllama(prompt string, images []string, model string) (*OllamaResponse, error) {
	return callOllamaWithSystem(prompt, images, model, fashionSystemPrompt)
}

func callOllamaWithSystem(prompt string, images []string, model, system string) (*OllamaResponse, error) {
	log.Printf("🤖 Calling Ollama model: %s", model)
	cleanImages := make([]string, len(images))
	for i, img := range images {
		cleanImages[i] = stripDataURIPrefix(img)
	}
	ollamaReq := OllamaRequest{
		Model:  model,
		Prompt: prompt,
		System: system,
		Images: cleanImages,
		Stream: false,
	}

	reqBody, _ := json.Marshal(ollamaReq)
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited (429)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &ollamaResp, nil
}

// ============================================================================
// OLLAMA CHAT API (for tool calling)
// ============================================================================

func callOllamaChat(messages []OllamaChatMessage, tools []OllamaToolDef, model string) (*OllamaChatResponse, error) {
	log.Printf("🤖 Calling Ollama chat model: %s (%d messages, %d tools)", model, len(messages), len(tools))

	req := OllamaChatRequest{
		Model:    model,
		Messages: messages,
		Tools:    tools,
		Stream:   false,
	}

	reqBody, _ := json.Marshal(req)
	resp, err := http.Post(ollamaURL+"/api/chat", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited (429)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp OllamaChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &chatResp, nil
}

func callOllamaChatWithFallback(messages []OllamaChatMessage, tools []OllamaToolDef, primaryModel, fallbackModel string) (*OllamaChatResponse, string, error) {
	resp, err := callOllamaChat(messages, tools, primaryModel)
	if err == nil {
		return resp, primaryModel, nil
	}

	if primaryModel == fallbackModel {
		return nil, primaryModel, err
	}

	log.Printf("⚠️  Chat %s failed (%v), falling back to %s", primaryModel, err, fallbackModel)
	resp, err = callOllamaChat(messages, tools, fallbackModel)
	if err != nil {
		return nil, fallbackModel, err
	}
	return resp, fallbackModel, nil
}

func runRemoveBgJob(ctx context.Context, job *AIJob, payload json.RawMessage) {
	var req RemoveBgRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		job.fail("Invalid payload: " + err.Error())
		return
	}

	if req.Image == "" {
		job.fail("Image is required")
		return
	}

	imageData := req.Image
	if idx := strings.Index(imageData, ","); idx != -1 {
		imageData = imageData[idx+1:]
	}

	imgBytes, err := base64.StdEncoding.DecodeString(imageData)
	if err != nil {
		job.fail("Invalid base64 image data")
		return
	}

	job.updateStatus("processing", "Removing background...")

	if ctx.Err() != nil {
		return
	}

	start := time.Now()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "image.png")
	if err != nil {
		job.fail("Failed to prepare request")
		return
	}
	part.Write(imgBytes)
	writer.Close()

	resp, err := http.Post(rembgURL+"/api/remove", writer.FormDataContentType(), &buf)
	if err != nil {
		job.fail("Background removal service not available: " + err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		job.fail(fmt.Sprintf("Background removal failed (status %d)", resp.StatusCode))
		return
	}

	processedBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		job.fail("Failed to read processed image")
		return
	}

	log.Printf("Background removed in %v (%d bytes -> %d bytes)", time.Since(start), len(imgBytes), len(processedBytes))

	processedBase64 := base64.StdEncoding.EncodeToString(processedBytes)
	job.complete(RemoveBgResponse{
		Image: "data:image/png;base64," + processedBase64,
	})
}

// rembgAvailable checks if the rembg service is running
func rembgAvailable() bool {
	resp, err := http.Get(rembgURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// removeBackgroundFromBytes calls the rembg service and returns the cleaned image bytes
func removeBackgroundFromBytes(imgBytes []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("file", "image.png")
	if err != nil {
		return nil, err
	}
	part.Write(imgBytes)
	writer.Close()

	resp, err := http.Post(rembgURL+"/api/remove", writer.FormDataContentType(), &buf)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rembg returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// uploadToGradio uploads a base64 image to a Gradio Space and returns a FileData reference.
// Newer Gradio (>=4) uses /gradio_api/upload and expects FileData with meta._type.
func uploadToGradio(spaceURL string, imageData string) (map[string]interface{}, error) {
	raw := imageData
	if idx := strings.Index(raw, ","); idx != -1 {
		raw = raw[idx+1:]
	}

	imgBytes, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	part, err := writer.CreateFormFile("files", "image.png")
	if err != nil {
		return nil, err
	}
	part.Write(imgBytes)
	writer.Close()

	// Try /gradio_api/upload first (Gradio >= 4), fall back to /upload
	uploadURL := spaceURL + "/gradio_api/upload"
	resp, err := http.Post(uploadURL, writer.FormDataContentType(), &buf)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusMethodNotAllowed {
		// Rebuild buffer for retry with legacy /upload
		buf.Reset()
		writer = multipart.NewWriter(&buf)
		part, _ = writer.CreateFormFile("files", "image.png")
		part.Write(imgBytes)
		writer.Close()

		resp.Body.Close()
		uploadURL = spaceURL + "/upload"
		resp, err = http.Post(uploadURL, writer.FormDataContentType(), &buf)
		if err != nil {
			return nil, fmt.Errorf("upload failed: %w", err)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upload returned %d: %s", resp.StatusCode, string(body))
	}

	var uploadResult []string
	if err := json.NewDecoder(resp.Body).Decode(&uploadResult); err != nil {
		return nil, fmt.Errorf("failed to parse upload response: %w", err)
	}

	if len(uploadResult) == 0 {
		return nil, fmt.Errorf("upload returned empty result")
	}

	// Return Gradio FileData format with path only (url causes issues with FASHN Space)
	filePath := uploadResult[0]
	return map[string]interface{}{
		"path": filePath,
		"meta": map[string]interface{}{
			"_type": "gradio.FileData",
		},
	}, nil
}

// mapClothType maps internal cloth type names to FASHN VTON categories.
// Internal: "upper", "lower", "overall" → FASHN: "tops", "bottoms", "one-pieces"
func mapClothType(clothType string) string {
	switch strings.ToLower(clothType) {
	case "lower", "bottom", "bottoms":
		return "bottoms"
	case "overall", "one-piece", "one-pieces", "dress", "full":
		return "one-pieces"
	default:
		return "tops"
	}
}

// callVtonBackend calls the FASHN VTON Gradio API (hmajid2301's Space).
// Uses Gradio /gradio_api/call pattern with FileData format.
func callVtonBackend(backendURL, personImage, garmentImage string) (string, error) {
	// Upload images to Gradio and get file references
	personRef, err := uploadToGradio(backendURL, personImage)
	if err != nil {
		return "", fmt.Errorf("person image upload: %w", err)
	}

	garmentRef, err := uploadToGradio(backendURL, garmentImage)
	if err != nil {
		return "", fmt.Errorf("garment image upload: %w", err)
	}

	// Build Gradio API call payload
	// /try_on params: person_image, garment_image, category, photo_type, steps, guidance, seed, segmentation_free
	callPayload := map[string]interface{}{
		"data": []interface{}{
			personRef,  // Person Image (FileData)
			garmentRef, // Garment Image (FileData)
			"tops",     // Category: tops, bottoms, one-pieces
			"flat-lay", // Photo Type: model or flat-lay
			50,         // Sampling Steps
			1.5,        // Guidance Scale
			42,         // Seed
			true,       // Segmentation Free
		},
	}

	payloadBytes, _ := json.Marshal(callPayload)

	log.Printf("🧥 Calling FASHN VTON /gradio_api/call/try_on")

	// POST to /gradio_api/call/try_on
	callResp, err := http.Post(backendURL+"/gradio_api/call/try_on", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("gradio call failed: %w", err)
	}
	defer callResp.Body.Close()

	if callResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(callResp.Body)
		return "", fmt.Errorf("gradio call returned %d: %s", callResp.StatusCode, string(body))
	}

	var callResult struct {
		EventID string `json:"event_id"`
	}
	if err := json.NewDecoder(callResp.Body).Decode(&callResult); err != nil {
		return "", fmt.Errorf("failed to parse event_id: %w", err)
	}

	if callResult.EventID == "" {
		return "", fmt.Errorf("no event_id returned")
	}

	log.Printf("🧥 FASHN VTON queued, event_id=%s", callResult.EventID)

	// GET results via SSE stream
	resultURL := fmt.Sprintf("%s/gradio_api/call/try_on/%s", backendURL, callResult.EventID)
	resultResp, err := http.Get(resultURL)
	if err != nil {
		return "", fmt.Errorf("result stream failed: %w", err)
	}
	defer resultResp.Body.Close()

	// Parse SSE stream for complete/error events with timeout
	scanner := bufio.NewScanner(resultResp.Body)
	scanner.Buffer(make([]byte, 0), 10*1024*1024)

	resultChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	go func() {
		var lastEvent string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "event: ") {
				lastEvent = strings.TrimPrefix(line, "event: ")
				log.Printf("🧥 SSE event: %s", lastEvent)
			} else if strings.HasPrefix(line, "data: ") {
				dataStr := strings.TrimPrefix(line, "data: ")

				// Only log first 200 chars of data to avoid spam
				logData := dataStr
				if len(logData) > 200 {
					logData = logData[:200] + "..."
				}
				log.Printf("🧥 SSE data (event=%s): %s", lastEvent, logData)

				if lastEvent == "complete" {
					var resultData []interface{}
					if err := json.Unmarshal([]byte(dataStr), &resultData); err != nil {
						errorChan <- fmt.Errorf("failed to parse result data: %w", err)
						return
					}

					if len(resultData) == 0 {
						errorChan <- fmt.Errorf("empty result data")
						return
					}

					// Result is FileData object with url or path
					resultMap, ok := resultData[0].(map[string]interface{})
					if !ok {
						errorChan <- fmt.Errorf("unexpected result format: %T", resultData[0])
						return
					}

					imageURL := ""
					if u, ok := resultMap["url"].(string); ok {
						imageURL = u
					} else if p, ok := resultMap["path"].(string); ok {
						imageURL = backendURL + "/file=" + p
					}

					if imageURL == "" {
						errorChan <- fmt.Errorf("no image URL in result: %v", resultMap)
						return
					}

					log.Printf("🧥 Downloading result image from: %s", imageURL)

					// Download the result image
					imgResp, err := http.Get(imageURL)
					if err != nil {
						errorChan <- fmt.Errorf("failed to download result: %w", err)
						return
					}
					defer imgResp.Body.Close()

					imgBytes, err := io.ReadAll(imgResp.Body)
					if err != nil {
						errorChan <- fmt.Errorf("failed to read result image: %w", err)
						return
					}

					log.Printf("🧥 FASHN VTON complete, downloaded %d bytes", len(imgBytes))
					resultChan <- "data:image/png;base64," + base64.StdEncoding.EncodeToString(imgBytes)
					return
				} else if lastEvent == "error" {
					// Don't fail on null errors - might be transient, wait for more events
					if dataStr == "null" || dataStr == "" {
						log.Printf("⚠ Received null/empty error from Gradio, waiting for more events...")
						continue
					}
					errorChan <- fmt.Errorf("gradio error: %s", dataStr)
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errorChan <- fmt.Errorf("scanner error: %w", err)
		} else {
			errorChan <- fmt.Errorf("no complete event received from Gradio")
		}
	}()

	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errorChan:
		return "", err
	case <-time.After(60 * time.Second):
		return "", fmt.Errorf("timeout waiting for Gradio response (60s)")
	}
}

// sendSSE writes a server-sent event to the response writer and flushes.
func sendSSE(w http.ResponseWriter, event, data string) {
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
}

// uploadFileToGradio uploads a local file to the Gradio server via POST /upload
// and returns the server-side path for use in predict calls.
func uploadFileToGradio(ctx context.Context, gradioURL, localPath string) (string, error) {
	f, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("files", "image.png")
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return "", fmt.Errorf("copy file data: %w", err)
	}
	w.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", gradioURL+"/upload", &buf)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed (status %d): %s", resp.StatusCode, string(body))
	}

	var paths []string
	if err := json.NewDecoder(resp.Body).Decode(&paths); err != nil {
		return "", fmt.Errorf("parse upload response: %w", err)
	}
	if len(paths) == 0 {
		return "", fmt.Errorf("upload returned no file paths")
	}

	return paths[0], nil
}

// createDummyMaskFile creates a temporary 1x1 black PNG file for use as an
// empty mask layer. CatVTON's AutoMasker activates when the mask is uniform
// (all same pixel value), so this dummy triggers proper DensePose+SCHP masking.
func createDummyMaskFile() (string, error) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{0, 0, 0, 255})

	tmpFile, err := os.CreateTemp("", "mask-*.png")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if err := png.Encode(tmpFile, img); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}
	return tmpFile.Name(), nil
}

func executeVTONStep(ctx context.Context, job *AIJob, personImage, garmentImage, clothType string, step, totalSteps int) (string, error) {
	if catvtonLocalURL == "" {
		return "", fmt.Errorf("virtual try-on unavailable (no local CatVTON service)")
	}

	stepDesc := ""
	if totalSteps > 1 {
		stepDesc = fmt.Sprintf(" (item %d/%d)", step, totalSteps)
	}

	job.updateStatus("processing", fmt.Sprintf("Processing %s try-on%s...", clothType, stepDesc))
	log.Printf("Using local CatVTON Gradio at %s for %s", catvtonLocalURL, clothType)

	// Save base64 images to temp files for upload to Gradio
	personFile, err := saveBase64ToTempFile(personImage, "person-*.png")
	if err != nil {
		return "", fmt.Errorf("failed to save person image: %w", err)
	}
	defer os.Remove(personFile)

	garmentFile, err := saveBase64ToTempFile(garmentImage, "garment-*.png")
	if err != nil {
		return "", fmt.Errorf("failed to save garment image: %w", err)
	}
	defer os.Remove(garmentFile)

	// Create dummy 1x1 black PNG for mask layer
	maskFile, err := createDummyMaskFile()
	if err != nil {
		return "", fmt.Errorf("failed to create mask file: %w", err)
	}
	defer os.Remove(maskFile)

	// Upload images to Gradio server
	job.updateStatus("processing", fmt.Sprintf("Uploading images%s...", stepDesc))

	personPath, err := uploadFileToGradio(ctx, catvtonLocalURL, personFile)
	if err != nil {
		return "", fmt.Errorf("failed to upload person image: %w", err)
	}

	garmentPath, err := uploadFileToGradio(ctx, catvtonLocalURL, garmentFile)
	if err != nil {
		return "", fmt.Errorf("failed to upload garment image: %w", err)
	}

	maskPath, err := uploadFileToGradio(ctx, catvtonLocalURL, maskFile)
	if err != nil {
		return "", fmt.Errorf("failed to upload mask image: %w", err)
	}

	// Build Gradio API payload
	gradioPayload := map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{
				"background": map[string]string{"path": personPath},
				"layers":     []interface{}{map[string]string{"path": maskPath}},
				"composite":  nil,
			},
			map[string]string{"path": garmentPath},
			clothType,
			30,  // steps
			2.5, // guidance
			42,  // seed
			"result only",
		},
	}

	payloadBytes, _ := json.Marshal(gradioPayload)

	gradioReq, err := http.NewRequestWithContext(ctx, "POST",
		catvtonLocalURL+"/api/submit_function", bytes.NewReader(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	gradioReq.Header.Set("Content-Type", "application/json")

	// Send periodic progress updates
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		elapsed := 0
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				elapsed += 10
				job.updateStatus("processing", fmt.Sprintf("Still processing %s%s... (%ds elapsed)", clothType, stepDesc, elapsed))
			}
		}
	}()

	start := time.Now()
	longClient := &http.Client{Timeout: 6 * time.Minute}
	gradioResp, err := longClient.Do(gradioReq)
	close(done)
	if err != nil {
		return "", fmt.Errorf("gradio request failed: %w", err)
	}
	defer gradioResp.Body.Close()

	body, _ := io.ReadAll(gradioResp.Body)

	if gradioResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gradio failed (status %d): %s", gradioResp.StatusCode, string(body))
	}

	var gradioResult struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &gradioResult); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(gradioResult.Data) == 0 {
		return "", fmt.Errorf("gradio returned empty data")
	}

	var resultFile struct {
		Path string `json:"path"`
		URL  string `json:"url"`
	}
	if err := json.Unmarshal(gradioResult.Data[0], &resultFile); err != nil {
		return "", fmt.Errorf("unexpected result format: %w", err)
	}

	var downloadURL string
	if resultFile.Path != "" {
		downloadURL = catvtonLocalURL + "/file=" + resultFile.Path
	} else if resultFile.URL != "" {
		downloadURL = resultFile.URL
		if strings.Contains(downloadURL, "localhost:7860") && !strings.Contains(catvtonLocalURL, "localhost") {
			downloadURL = strings.Replace(downloadURL, "http://localhost:7860", catvtonLocalURL, 1)
		}
	} else {
		return "", fmt.Errorf("result has no path or URL")
	}

	downloadReq, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create download request: %w", err)
	}
	downloadResp, err := http.DefaultClient.Do(downloadReq)
	if err != nil {
		return "", fmt.Errorf("failed to download result: %w", err)
	}
	defer downloadResp.Body.Close()

	if downloadResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download result (status %d)", downloadResp.StatusCode)
	}

	imageData, err := io.ReadAll(downloadResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read result image: %w", err)
	}

	// Save debug copy
	saveDir := filepath.Join("..", "images", "generated")
	if err := os.MkdirAll(saveDir, 0755); err == nil {
		savePath := filepath.Join(saveDir, fmt.Sprintf("vton-%s-step%d.webp", job.ID, step))
		os.WriteFile(savePath, imageData, 0644)
	}

	contentType := downloadResp.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = "image/webp"
	}

	log.Printf("CatVTON step %d complete in %v", step, time.Since(start))
	return fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(imageData)), nil
}

func runTryOnJob(ctx context.Context, job *AIJob, payload json.RawMessage) {
	var req TryOnRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		job.fail("Invalid payload: " + err.Error())
		return
	}

	// Normalize input to list of garments
	var garments []TryOnItem
	if len(req.Garments) > 0 {
		garments = req.Garments
	} else {
		// Legacy single item
		if req.PersonImage == "" || req.GarmentImage == "" {
			job.fail("person_image and garment_image are required")
			return
		}
		clothType := req.ClothType
		if clothType == "" {
			clothType = "upper"
		}
		garments = []TryOnItem{{Image: req.GarmentImage, ClothType: clothType}}
	}

	// Sort garments: process 'lower' (trousers) first to establish base
	sort.SliceStable(garments, func(i, j int) bool {
		// "lower" < "upper"/"overall"
		if garments[i].ClothType == "lower" && garments[j].ClothType != "lower" {
			return true
		}
		return false
	})

	currentPersonImage := req.PersonImage

	// Process each garment sequentially
	for i, garment := range garments {
		if garment.ClothType == "shoes" {
			// Skip shoes or handle separately
			continue
		}

		resultImage, err := executeVTONStep(ctx, job, currentPersonImage, garment.Image, garment.ClothType, i+1, len(garments))
		if err != nil {
			job.fail(fmt.Sprintf("Failed at step %d (%s): %v", i+1, garment.ClothType, err))
			return
		}
		currentPersonImage = resultImage
	}

	job.complete(&TryOnResponse{Image: currentPersonImage})
}

func saveBase64ToTempFile(data, pattern string) (string, error) {
	// Strip data URI prefix if present
	if strings.Contains(data, ",") {
		data = strings.Split(data, ",")[1]
	}

	imageData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(imageData); err != nil {
		return "", err
	}

	return tmpFile.Name(), nil
}

// runShoesTryOn handles shoe virtual try-on via the LaDI-VTON-Shoes service
func runShoesTryOn(ctx context.Context, job *AIJob, req TryOnRequest) {
	job.updateStatus("processing", "👟 Processing shoe try-on...")

	if ladivtonShoesURL == "" {
		job.fail("Shoe try-on unavailable (no service configured)")
		return
	}

	start := time.Now()
	log.Printf("👟 Sending to LaDI-VTON-Shoes at %s", ladivtonShoesURL)

	payload, _ := json.Marshal(map[string]string{
		"person_image": req.PersonImage,
		"shoe_image":   req.GarmentImage,
	})

	shoesReq, err := http.NewRequestWithContext(ctx, "POST", ladivtonShoesURL+"/api/tryon", bytes.NewReader(payload))
	if err != nil {
		job.fail(fmt.Sprintf("Failed to create request: %v", err))
		return
	}
	shoesReq.Header.Set("Content-Type", "application/json")

	// Send periodic progress updates while waiting
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		elapsed := 0
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				elapsed += 10
				job.updateStatus("processing", fmt.Sprintf("👟 Still processing shoes... (%ds elapsed)", elapsed))
			}
		}
	}()

	resp, err := http.DefaultClient.Do(shoesReq)
	close(done)

	if err != nil {
		log.Printf("❌ LaDI-VTON-Shoes request failed: %v", err)
		if ctx.Err() != nil {
			return
		}
		job.fail(fmt.Sprintf("Shoe try-on failed: %v", err))
		return
	}
	defer resp.Body.Close()

	log.Printf("👟 LaDI-VTON-Shoes response: status=%d", resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ LaDI-VTON-Shoes error response: %s", string(body))
		job.fail(fmt.Sprintf("Shoe try-on failed (status %d): %s", resp.StatusCode, string(body)))
		return
	}

	var result TryOnResponse
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("❌ Failed to parse shoe response: %v", err)
		job.fail("Failed to parse shoe result: " + err.Error())
		return
	}

	log.Printf("👟 LaDI-VTON-Shoes try-on complete in %v", time.Since(start))
	job.complete(&result)
}

func runSuggestOutfitJob(ctx context.Context, job *AIJob, payload json.RawMessage) {
	var req SuggestOutfitRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		job.fail("Invalid payload: " + err.Error())
		return
	}

	if len(req.Items) == 0 {
		job.fail("At least one item required")
		return
	}

	job.updateStatus("processing", "AI is selecting outfit...")

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

Select items for a complete outfit. You MUST include ALL of these:
- ONE shirt or t-shirt (REQUIRED - a sweater or cardigan is NOT a shirt, it goes on top of a shirt)
- ONE bottom (pants, jeans, skirt, shorts)
- ONE pair of shoes (REQUIRED - every outfit needs shoes)
- OPTIONAL: One outerwear layer (jacket, coat, cardigan, sweater) - if weather requires it, this goes OVER the shirt
- OPTIONAL: 1-2 accessories (watch, belt, scarf, hat, bag)

An outfit without shoes is INCOMPLETE. Always include shoes.

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

	if ctx.Err() != nil {
		return
	}

	start := time.Now()
	ollamaResp, usedModel, err := callOllamaWithFallback(prompt, nil, cloudTextModel, textModel)
	if err != nil {
		job.fail("AI request failed: " + err.Error())
		return
	}

	log.Printf("Suggest outfit job %s: %v (model: %s)", job.ID[:8], time.Since(start), usedModel)

	result := parseSuggestOutfitResponse(ollamaResp.Response)
	result.RawResponse = ollamaResp.Response
	job.complete(result)
}

func runColorMatchJob(ctx context.Context, job *AIJob, payload json.RawMessage) {
	var req ColorMatchRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		job.fail("Invalid payload: " + err.Error())
		return
	}

	if len(req.Items) < 2 {
		job.fail("At least 2 items required for color matching")
		return
	}

	job.updateStatus("processing", "Analyzing color coordination...")

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

	if ctx.Err() != nil {
		return
	}

	ollamaResp, _, err := callOllamaWithFallback(prompt, nil, cloudTextModel, textModel)
	if err != nil {
		job.fail("AI request failed: " + err.Error())
		return
	}

	result := parseColorMatchResponse(ollamaResp.Response)
	result.RawResponse = ollamaResp.Response
	job.complete(result)
}

func runStyleMatchJob(ctx context.Context, job *AIJob, payload json.RawMessage) {
	var req StyleMatchRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		job.fail("Invalid payload: " + err.Error())
		return
	}

	if req.BaseItem.ID == 0 {
		job.fail("Base item required")
		return
	}

	if len(req.Items) == 0 {
		job.fail("Items to match required")
		return
	}

	job.updateStatus("processing", "Finding matching items...")

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

	if ctx.Err() != nil {
		return
	}

	ollamaResp, _, err := callOllamaWithFallback(prompt, nil, cloudTextModel, textModel)
	if err != nil {
		job.fail("AI request failed: " + err.Error())
		return
	}

	result := parseStyleMatchResponse(ollamaResp.Response)
	result.RawResponse = ollamaResp.Response
	job.complete(result)
}

func runRateOutfitJob(ctx context.Context, job *AIJob, payload json.RawMessage) {
	var req RateOutfitRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		job.fail("Invalid payload: " + err.Error())
		return
	}

	if len(req.Items) == 0 {
		job.fail("At least one item required")
		return
	}

	job.updateStatus("processing", "AI is rating your outfit...")

	itemsList := ""
	for _, item := range req.Items {
		tags := strings.Join(item.Tags, ", ")
		desc := ""
		if item.Description != "" {
			desc = fmt.Sprintf(" — %s", item.Description)
		}
		itemsList += fmt.Sprintf("- %s (%s, %s) [%s]%s\n", item.Name, item.Category, item.Color, tags, desc)
	}

	prompt := fmt.Sprintf(`You are rating an outfit that the user has put together. The user intends to wear ALL of these items together as one outfit.

Items in this outfit:
%s

Rate this outfit from 0-10 based on how well these items work TOGETHER.

Rules:
- The user IS wearing all these items together. Do not suggest removing items or wearing items "on their own".
- Only reference garment features that are explicitly stated in the descriptions. Do not invent or assume details like neckline type, pattern, or fabric unless described.
- Focus on: color harmony, style cohesion, whether the layering order makes sense, and overall balance.
- Improvements should suggest what ADDITIONAL items or SWAPS from a wardrobe could enhance the outfit, not removing items already chosen.

Respond ONLY with valid JSON:
{
  "rating": 7,
  "feedback": "1-2 sentence overall assessment",
  "strengths": ["strength 1", "strength 2"],
  "improvements": ["actionable suggestion 1", "actionable suggestion 2"],
  "color_score": 85,
  "color_analysis": "1 sentence about how the colors work together"
}`, itemsList)

	if ctx.Err() != nil {
		return
	}

	ollamaResp, _, err := callOllamaWithFallback(prompt, nil, cloudTextModel, textModel)
	if err != nil {
		job.fail("AI request failed: " + err.Error())
		return
	}

	result := parseRateOutfitResponse(ollamaResp.Response)
	result.RawResponse = ollamaResp.Response
	job.complete(result)
}

func runWardrobeGapsJob(ctx context.Context, job *AIJob, payload json.RawMessage) {
	var req WardrobeGapsRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		job.fail("Invalid payload: " + err.Error())
		return
	}

	if len(req.Items) == 0 {
		job.fail("At least one item required")
		return
	}

	job.updateStatus("processing", "Analyzing your wardrobe...")

	categoryCount := make(map[string]int)
	itemsList := ""

	for _, item := range req.Items {
		categoryCount[item.Category]++
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

	if ctx.Err() != nil {
		return
	}

	ollamaResp, _, err := callOllamaWithFallback(prompt, nil, cloudTextModel, textModel)
	if err != nil {
		job.fail("AI request failed: " + err.Error())
		return
	}

	result := parseWardrobeGapsResponse(ollamaResp.Response)
	result.RawResponse = ollamaResp.Response
	job.complete(result)
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
		Rating        int      `json:"rating"`
		Feedback      string   `json:"feedback"`
		Strengths     []string `json:"strengths"`
		Improvements  []string `json:"improvements"`
		ColorScore    int      `json:"color_score"`
		ColorAnalysis string   `json:"color_analysis"`
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
		if parsed.ColorScore > 0 {
			result.ColorScore = parsed.ColorScore
		}
		if parsed.ColorAnalysis != "" {
			result.ColorAnalysis = parsed.ColorAnalysis
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

// ============================================================================
// AI CHAT WITH TOOL CALLING
// ============================================================================

const maxToolIterations = 10

func getWardrobeTools() []OllamaToolDef {
	return []OllamaToolDef{
		{
			Type: "function",
			Function: OllamaFunctionDef{
				Name:        "list_items",
				Description: "List clothing items in the wardrobe. Can filter by category, color, season, or brand.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"category": map[string]interface{}{"type": "string", "description": "Filter by category: Tops, Bottoms, Shoes, Outerwear, Accessories"},
						"color":    map[string]interface{}{"type": "string", "description": "Filter by color"},
						"season":   map[string]interface{}{"type": "string", "description": "Filter by season: All Season, Spring, Summer, Fall, Winter"},
						"brand":    map[string]interface{}{"type": "string", "description": "Filter by brand name"},
						"limit":    map[string]interface{}{"type": "integer", "description": "Max items to return (default 20)"},
					},
				},
			},
		},
		{
			Type: "function",
			Function: OllamaFunctionDef{
				Name:        "get_item",
				Description: "Get full details of a specific clothing item by its ID.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"item_id": map[string]interface{}{"type": "integer", "description": "The item ID"},
					},
					"required": []string{"item_id"},
				},
			},
		},
		{
			Type: "function",
			Function: OllamaFunctionDef{
				Name:        "list_outfits",
				Description: "List saved outfits with their item IDs and wear counts.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"limit": map[string]interface{}{"type": "integer", "description": "Max outfits to return (default 20)"},
					},
				},
			},
		},
		{
			Type: "function",
			Function: OllamaFunctionDef{
				Name:        "get_outfit",
				Description: "Get full details of a specific outfit including all its items.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"outfit_id": map[string]interface{}{"type": "integer", "description": "The outfit ID"},
					},
					"required": []string{"outfit_id"},
				},
			},
		},
		{
			Type: "function",
			Function: OllamaFunctionDef{
				Name:        "get_wardrobe_stats",
				Description: "Get wardrobe statistics: total items, counts by category, by color, and by season.",
				Parameters: map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
				},
			},
		},
		{
			Type: "function",
			Function: OllamaFunctionDef{
				Name:        "get_wear_history",
				Description: "Get wear history — when items or outfits were worn.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"item_id":   map[string]interface{}{"type": "integer", "description": "Filter by item ID"},
						"outfit_id": map[string]interface{}{"type": "integer", "description": "Filter by outfit ID"},
						"days_back": map[string]interface{}{"type": "integer", "description": "Only show history from last N days (default 30)"},
					},
				},
			},
		},
		{
			Type: "function",
			Function: OllamaFunctionDef{
				Name:        "get_most_worn",
				Description: "Get the most frequently worn clothing items.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"limit": map[string]interface{}{"type": "integer", "description": "Number of items to return (default 10)"},
					},
				},
			},
		},
		{
			Type: "function",
			Function: OllamaFunctionDef{
				Name:        "get_least_worn",
				Description: "Get the least frequently worn clothing items (items that need more love).",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"limit": map[string]interface{}{"type": "integer", "description": "Number of items to return (default 10)"},
					},
				},
			},
		},
		{
			Type: "function",
			Function: OllamaFunctionDef{
				Name:        "get_calendar_events",
				Description: "Get calendar events showing what outfits were planned/worn on which dates, including weather and location.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"start_date": map[string]interface{}{"type": "string", "description": "Start date (YYYY-MM-DD)"},
						"end_date":   map[string]interface{}{"type": "string", "description": "End date (YYYY-MM-DD)"},
					},
				},
			},
		},
		{
			Type: "function",
			Function: OllamaFunctionDef{
				Name:        "search_items",
				Description: "Search for clothing items by name, brand, or description text.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"query": map[string]interface{}{"type": "string", "description": "Search text to match against item names, brands, and descriptions"},
					},
					"required": []string{"query"},
				},
			},
		},
	}
}

func intFromArgs(args map[string]interface{}, key string, def int) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return def
}

func int64FromArgs(args map[string]interface{}, key string) *int64 {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			i := int64(n)
			return &i
		case int:
			i := int64(n)
			return &i
		}
	}
	return nil
}

func stringFromArgs(args map[string]interface{}, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}

func executeToolCall(ctx context.Context, userID int64, tc OllamaToolCall) (string, error) {
	var args map[string]interface{}
	if err := json.Unmarshal(tc.Function.Arguments, &args); err != nil {
		return fmt.Sprintf(`{"error": "invalid arguments: %s"}`, err.Error()), nil
	}

	var result interface{}
	var err error

	switch tc.Function.Name {
	case "list_items":
		result, err = toolListItems(ctx, userID,
			stringFromArgs(args, "category"),
			stringFromArgs(args, "color"),
			stringFromArgs(args, "season"),
			stringFromArgs(args, "brand"),
			intFromArgs(args, "limit", 20))

	case "get_item":
		itemID := int64FromArgs(args, "item_id")
		if itemID == nil {
			return `{"error": "item_id is required"}`, nil
		}
		result, err = toolGetItem(ctx, userID, *itemID)

	case "list_outfits":
		result, err = toolListOutfits(ctx, userID, intFromArgs(args, "limit", 20))

	case "get_outfit":
		outfitID := int64FromArgs(args, "outfit_id")
		if outfitID == nil {
			return `{"error": "outfit_id is required"}`, nil
		}
		result, err = toolGetOutfit(ctx, userID, *outfitID)

	case "get_wardrobe_stats":
		stats, dbErr := toolGetWardrobeStats(ctx, userID)
		if dbErr != nil {
			err = dbErr
		} else {
			result, err = json.Marshal(stats)
			if err == nil {
				result = string(result.([]byte))
			}
		}

	case "get_wear_history":
		result, err = toolGetWearHistory(ctx, userID,
			int64FromArgs(args, "item_id"),
			int64FromArgs(args, "outfit_id"),
			intFromArgs(args, "days_back", 30))

	case "get_most_worn":
		items, dbErr := toolGetMostWorn(ctx, userID, intFromArgs(args, "limit", 10))
		if dbErr != nil {
			err = dbErr
		} else {
			result, err = json.Marshal(items)
			if err == nil {
				result = string(result.([]byte))
			}
		}

	case "get_least_worn":
		items, dbErr := toolGetLeastWorn(ctx, userID, intFromArgs(args, "limit", 10))
		if dbErr != nil {
			err = dbErr
		} else {
			result, err = json.Marshal(items)
			if err == nil {
				result = string(result.([]byte))
			}
		}

	case "get_calendar_events":
		result, err = toolGetCalendarEvents(ctx, userID,
			stringFromArgs(args, "start_date"),
			stringFromArgs(args, "end_date"))

	case "search_items":
		query := stringFromArgs(args, "query")
		if query == "" {
			return `{"error": "query is required"}`, nil
		}
		result, err = toolSearchItems(ctx, userID, query)

	default:
		return fmt.Sprintf(`{"error": "unknown tool: %s"}`, tc.Function.Name), nil
	}

	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error()), nil
	}

	resultJSON, _ := json.Marshal(result)
	return string(resultJSON), nil
}

func runToolCallingLoop(ctx context.Context, messages []OllamaChatMessage, tools []OllamaToolDef, model, fallbackModel string, userID int64, statusFn func(string)) (*OllamaChatResponse, int, error) {
	totalToolCalls := 0

	for i := 0; i < maxToolIterations; i++ {
		if ctx.Err() != nil {
			return nil, totalToolCalls, ctx.Err()
		}

		resp, _, err := callOllamaChatWithFallback(messages, tools, model, fallbackModel)
		if err != nil {
			return nil, totalToolCalls, err
		}

		// No tool calls — done
		if len(resp.Message.ToolCalls) == 0 {
			return resp, totalToolCalls, nil
		}

		// Append assistant message (with tool_calls) to history
		messages = append(messages, resp.Message)

		// Execute each tool call and append results
		for _, tc := range resp.Message.ToolCalls {
			totalToolCalls++
			log.Printf("🔧 Tool call: %s", tc.Function.Name)
			if statusFn != nil {
				statusFn(fmt.Sprintf("Querying: %s...", tc.Function.Name))
			}

			result, err := executeToolCall(ctx, userID, tc)
			if err != nil {
				result = fmt.Sprintf(`{"error": "%s"}`, err.Error())
			}

			messages = append(messages, OllamaChatMessage{
				Role:    "tool",
				Content: result,
			})
		}
	}

	return nil, totalToolCalls, fmt.Errorf("exceeded max tool iterations (%d)", maxToolIterations)
}

func runChatJob(ctx context.Context, job *AIJob, payload json.RawMessage) {
	var req ChatRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		job.fail("Invalid payload: " + err.Error())
		return
	}

	if req.Message == "" {
		job.fail("Message is required")
		return
	}

	userID := int64(1) // TODO: auth

	job.updateStatus("processing", "Thinking...")

	messages := []OllamaChatMessage{
		{
			Role: "system",
			Content: `You are a wardrobe assistant for GoThreads. You help users with outfit suggestions, wardrobe analysis, and style advice.

You have access to tools that let you query the user's wardrobe database. ALWAYS use the tools to look up actual wardrobe data — do not make assumptions about what items the user owns.

When suggesting outfits:
- Reference items by their actual names and IDs from the database
- Explain why the items work well together (color coordination, style, occasion)
- Consider wear history to suggest underutilized items

Keep responses concise and conversational. Use the tools proactively — if the user asks about their wardrobe, query it first before answering.`,
		},
	}

	// Append conversation history if provided
	for _, msg := range req.History {
		messages = append(messages, msg)
	}

	// Append the new user message
	messages = append(messages, OllamaChatMessage{
		Role:    "user",
		Content: req.Message,
	})

	tools := getWardrobeTools()

	if ctx.Err() != nil {
		return
	}

	start := time.Now()
	resp, toolsUsed, err := runToolCallingLoop(ctx, messages, tools, cloudTextModel, textModel, userID,
		func(status string) { job.updateStatus("processing", status) })
	if err != nil {
		job.fail("AI request failed: " + err.Error())
		return
	}

	log.Printf("🤖 Chat complete in %v: %d tool calls made", time.Since(start), toolsUsed)

	job.complete(ChatResponse{
		Message:   resp.Message.Content,
		ToolsUsed: toolsUsed,
	})
}

func runDetectOutfitItemsJob(ctx context.Context, job *AIJob, payload json.RawMessage) {
	var req DetectOutfitItemsRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		job.fail("Invalid payload: " + err.Error())
		return
	}

	if req.Image == "" {
		job.fail("Image is required")
		return
	}

	job.updateStatus("processing", "🔍 Analyzing image for clothing items...")

	prompt := `Analyze this image and detect all visible clothing items and accessories. For each item, provide:
1. Category (tops, bottoms, shoes, accessories, outerwear, dress, one-piece)
2. Detailed description (e.g., "light blue denim jacket with buttons", "black leather ankle boots")
3. Primary color
4. Confidence level (high/medium/low)

Return ONLY valid JSON in this exact format:
{
  "items": [
    {
      "category": "tops",
      "description": "light blue denim jacket",
      "color": "blue",
      "confidence": "high"
    }
  ],
  "count": 1,
  "description": "Brief summary of the outfit"
}`

	if ctx.Err() != nil {
		return
	}

	start := time.Now()
	ollamaResp, usedModel, err := callOllamaWithFallback(prompt, []string{req.Image}, spatialModel, visionModel)
	if err != nil {
		job.fail("AI request failed: " + err.Error())
		return
	}

	log.Printf("Detect outfit items job %s: %v (model: %s)", job.ID[:8], time.Since(start), usedModel)

	result := parseDetectOutfitItemsResponse(ollamaResp.Response)
	result.RawResponse = ollamaResp.Response
	job.complete(result)
}

func parseDetectOutfitItemsResponse(response string) *DetectOutfitItemsResponse {
	// Try to extract JSON from the response
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start == -1 || end == -1 {
		return &DetectOutfitItemsResponse{
			Items:       []DetectedItem{},
			Count:       0,
			Description: "Could not parse AI response",
		}
	}

	jsonStr := response[start : end+1]

	var result DetectOutfitItemsResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		log.Printf("Failed to parse detect response: %v", err)
		return &DetectOutfitItemsResponse{
			Items:       []DetectedItem{},
			Count:       0,
			Description: "Invalid AI response format",
		}
	}

	return &result
}

// handleScrapeURL scrapes product info from a URL using simple OG tag parsing
func handleScrapeURL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.URL == "" {
		writeError(w, http.StatusBadRequest, "URL is required")
		return
	}

	log.Printf("🕸️ Scraping URL: %s", req.URL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(req.URL)
	if err != nil {
		log.Printf("❌ Failed to fetch URL %s: %v", req.URL, err)
		writeError(w, http.StatusBadGateway, "Failed to fetch URL: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("❌ Failed to fetch URL %s: Status %d", req.URL, resp.StatusCode)
		writeError(w, http.StatusBadGateway, fmt.Sprintf("Failed to fetch URL (Status %d)", resp.StatusCode))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("❌ Failed to read response body from %s: %v", req.URL, err)
		writeError(w, http.StatusInternalServerError, "Failed to read response body")
		return
	}

	htmlContent := string(body)

	extractTag := func(property string) string {
		// Look for <meta property="property" content="...">
		re := regexp.MustCompile(fmt.Sprintf(`<meta\s+(?:property|name)=["']%s["']\s+content=["']([^"']+)["']`, regexp.QuoteMeta(property)))
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			return matches[1]
		}
		// Try reverse order: content="..." property="..."
		re = regexp.MustCompile(fmt.Sprintf(`<meta\s+content=["']([^"']+)["']\s+(?:property|name)=["']%s["']`, regexp.QuoteMeta(property)))
		matches = re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			return matches[1]
		}
		return ""
	}

	title := extractTag("og:title")
	if title == "" {
		// Fallback to <title>
		re := regexp.MustCompile(`<title>([^<]+)</title>`)
		matches := re.FindStringSubmatch(htmlContent)
		if len(matches) > 1 {
			title = matches[1]
		}
	}

	description := extractTag("og:description")
	siteName := extractTag("og:site_name")

	// Extract all potential product images
	var candidateImages []string
	seenImages := make(map[string]bool)

	// Helper to add image if new
	addImage := func(url string) {
		if url == "" || seenImages[url] {
			return
		}
		// Fix relative URLs
		if strings.HasPrefix(url, "//") {
			url = "https:" + url
		} else if strings.HasPrefix(url, "/") {
			// simple base URL logic
			baseURLParts := strings.Split(req.URL, "/")
			if len(baseURLParts) > 2 {
				url = baseURLParts[0] + "//" + baseURLParts[2] + url
			}
		}

		seenImages[url] = true
		candidateImages = append(candidateImages, url)
	}

	// 1. og:image (high priority)
	if ogImg := extractTag("og:image"); ogImg != "" {
		addImage(ogImg)
	}

	// 2. JSON-LD images (often high quality product shots)
	// Simple regex for "image": "url" or "image": ["url", ...] inside script tags
	jsonLdRe := regexp.MustCompile(`<script type="application/ld\+json">([\s\S]*?)</script>`)
	ldMatches := jsonLdRe.FindAllStringSubmatch(htmlContent, -1)
	for _, match := range ldMatches {
		if len(match) > 1 {
			// Look for "image": "..." or "image": [...]
			// Very naive regex, but might catch some
			imgRe := regexp.MustCompile(`"image"\s*:\s*(?:\[\s*)?"([^"]+)"`)
			imgMatches := imgRe.FindAllStringSubmatch(match[1], -1)
			for _, im := range imgMatches {
				if len(im) > 1 {
					addImage(im[1])
				}
			}
		}
	}

	// 3. Regular <img> tags (limit to first 10 large-ish looking ones to save AI tokens)
	imgTagRe := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`)
	imgTagMatches := imgTagRe.FindAllStringSubmatch(htmlContent, -1)
	for i, match := range imgTagMatches {
		if i > 5 {
			break
		} // limit
		if len(match) > 1 {
			addImage(match[1])
		}
	}

	// Limit candidates
	if len(candidateImages) > 5 {
		candidateImages = candidateImages[:5]
	}

	selectedImageURL := ""

	// If we have multiple images and AI is ready, ask AI to pick the best "no model" one
	if len(candidateImages) > 1 && isModelAvailable(visionModel) {
		log.Printf("🤖 Asking AI to pick best product image from %d candidates...", len(candidateImages))

		for i, imgUrl := range candidateImages {
			if i >= 3 {
				break
			} // Check top 3 only

			log.Printf("   Checking candidate %d: %s", i, imgUrl)
			// Download to memory
			resp, err := client.Get(imgUrl)
			if err != nil {
				continue
			}

			data, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}

			// Convert to base64
			b64 := base64.StdEncoding.EncodeToString(data)

			// Ask LLaVA: "Is this a plain product shot without a model?"
			prompt := `Does this image show a clothing item on a plain background or ghost mannequin, WITHOUT a human model visible? 
			Respond with YES or NO.`

			ollamaResp, err := callOllama(prompt, []string{b64}, visionModel)
			if err == nil {
				ans := strings.ToUpper(strings.TrimSpace(ollamaResp.Response))
				log.Printf("   AI says: %s", ans)
				if strings.Contains(ans, "YES") {
					selectedImageURL = imgUrl
					break // Found one!
				}
			}
		}
	}

	// Fallback to first candidate (usually og:image)
	if selectedImageURL == "" && len(candidateImages) > 0 {
		selectedImageURL = candidateImages[0]
	}

	if selectedImageURL == "" {
		log.Printf("⚠️  Scrape warning: No suitable image found for %s", req.URL)
		writeError(w, http.StatusNotFound, "No product image found")
		return
	}

	// Download and save the SELECTED image
	log.Printf("⬇️ Downloading selected image: %s", selectedImageURL)
	finalImgResp, err := client.Get(selectedImageURL)
	if err != nil {
		writeError(w, http.StatusBadGateway, "Failed to download image: "+err.Error())
		return
	}
	defer finalImgResp.Body.Close()

	finalImgBytes, err := io.ReadAll(finalImgResp.Body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to read image data")
		return
	}

	// Generate filename
	ext := ".jpg"
	if ct := http.DetectContentType(finalImgBytes); ct == "image/png" {
		ext = ".png"
	} else if ct == "image/webp" {
		ext = ".webp"
	}

	filename := fmt.Sprintf("scraped_%d%s", time.Now().UnixNano(), ext)
	localDir := "../images"
	if err := os.MkdirAll(localDir, 0755); err != nil {
		log.Printf("⚠️  Failed to create images directory: %v", err)
	}
	localPath := filepath.Join(localDir, filename) // Save to mockups/images

	if err := os.WriteFile(localPath, finalImgBytes, 0644); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save image locally: "+err.Error())
		return
	}

	// Return local URL
	localURL := fmt.Sprintf("http://localhost:%s/images/%s", getEnv("PORT", "8556"), filename)

	log.Printf("✅ Scraped & Saved: %s -> %s", req.URL, localURL)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"title":        title,
		"image_url":    localURL, // Return local URL
		"original_url": req.URL,
		"description":  description,
		"site_name":    siteName,
	})
}

func runScrapeJob(ctx context.Context, job *AIJob, payload json.RawMessage) {
	var req ScrapeRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		job.fail("Invalid payload: " + err.Error())
		return
	}

	job.updateStatus("processing", fmt.Sprintf("Scraping URL: %s", req.URL))

	// Use go-colly for robust scraping
	scraper := NewCollyScraper()
	result, err := scraper.Scrape(ctx, req.URL)
	if err != nil {
		job.fail("Failed to scrape URL: " + err.Error())
		return
	}

	title := result.Title
	description := result.Description
	siteName := result.SiteName
	var category, color, brand, price string
	brand = result.Brand
	price = result.Price

	// Convert image URLs to candidate list
	var candidateImages []string
	seenImages := make(map[string]bool)
	for _, imgURL := range result.ImageURLs {
		if imgURL == "" || seenImages[imgURL] {
			continue
		}
		// Fix relative URLs
		if strings.HasPrefix(imgURL, "//") {
			imgURL = "https:" + imgURL
		} else if strings.HasPrefix(imgURL, "/") {
			parsedURL, _ := url.Parse(req.URL)
			if parsedURL != nil {
				imgURL = fmt.Sprintf("https://%s%s", parsedURL.Hostname(), imgURL)
			}
		}
		seenImages[imgURL] = true
		candidateImages = append(candidateImages, imgURL)
	}

	// Limit candidates
	if len(candidateImages) > 5 {
		candidateImages = candidateImages[:5]
	}

	// AI Extraction from page text (if model available and data is incomplete)
	if isModelAvailable(textModel) || isModelAvailable(cloudTextModel) {
		job.updateStatus("processing", "AI analyzing page content...")

		// Fetch HTML content for text analysis
		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(req.URL)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			htmlContent := string(body)

			// Simple HTML text extraction
			cleanText := htmlContent
			reScript := regexp.MustCompile(`(?s)<script.*?>.*?</script>`)
			cleanText = reScript.ReplaceAllString(cleanText, "")
			reStyle := regexp.MustCompile(`(?s)<style.*?>.*?</style>`)
			cleanText = reStyle.ReplaceAllString(cleanText, "")
			reTags := regexp.MustCompile(`<[^>]*>`)
			cleanText = reTags.ReplaceAllString(cleanText, " ")
			reSpace := regexp.MustCompile(`\s+`)
			cleanText = reSpace.ReplaceAllString(cleanText, " ")
			if len(cleanText) > 4000 {
				cleanText = cleanText[:4000]
			}

			prompt := fmt.Sprintf(`Analyze this product page text and extract details.
			
			Text: "%s"
			
			Return ONLY JSON:
			{
				"title": "Product Name",
				"description": "Short description",
				"category": "Tops/Bottoms/Shoes/Outerwear/Accessories/Dresses",
				"color": "Main color",
				"brand": "Brand Name",
				"price": "Price if found (e.g. $50)"
			}`, cleanText)

			model := textModel
			if ollamaCloud {
				model = cloudTextModel
			}

			if aiResp, err := callOllama(prompt, nil, model); err == nil {
				var extracted struct {
					Title       string `json:"title"`
					Description string `json:"description"`
					Category    string `json:"category"`
					Color       string `json:"color"`
					Brand       string `json:"brand"`
					Price       string `json:"price"`
				}

				jsonStr := aiResp.Response
				if idx := strings.Index(jsonStr, "{"); idx != -1 {
					jsonStr = jsonStr[idx:]
					if lastIdx := strings.LastIndex(jsonStr, "}"); lastIdx != -1 {
						jsonStr = jsonStr[:lastIdx+1]
						if err := json.Unmarshal([]byte(jsonStr), &extracted); err == nil {
							log.Printf("🤖 AI Extracted: %+v", extracted)

							if title == "" || strings.Contains(title, "http") {
								title = extracted.Title
							}
							if description == "" {
								description = extracted.Description
							}
							if siteName == "" {
								siteName = extracted.Brand
							}

							category = extracted.Category
							color = extracted.Color
							if brand == "" {
								brand = extracted.Brand
							}
							if price == "" {
								price = extracted.Price
							}
						}
					}
				}
			}
		}
	}

	selectedImageURL := ""

	// AI Selection for best image
	if len(candidateImages) > 1 && isModelAvailable(visionModel) {
		job.updateStatus("processing", "AI selecting best product image...")
		log.Printf("🤖 Asking AI to pick best product image from %d candidates...", len(candidateImages))

		client := &http.Client{Timeout: 30 * time.Second}
		for i, imgUrl := range candidateImages {
			if i >= 3 {
				break
			}

			if ctx.Err() != nil {
				return
			}

			// Download to memory
			imgResp, err := client.Get(imgUrl)
			if err != nil {
				continue
			}

			data, err := io.ReadAll(imgResp.Body)
			imgResp.Body.Close()
			if err != nil {
				continue
			}

			b64 := base64.StdEncoding.EncodeToString(data)
			prompt := `Does this image show a clothing item on a plain background or ghost mannequin, WITHOUT a human model visible? Respond with YES or NO.`

			ollamaResp, err := callOllama(prompt, []string{b64}, visionModel)
			if err == nil {
				ans := strings.ToUpper(strings.TrimSpace(ollamaResp.Response))
				if strings.Contains(ans, "YES") {
					selectedImageURL = imgUrl
					break
				}
			}
		}
	}

	if selectedImageURL == "" && len(candidateImages) > 0 {
		selectedImageURL = candidateImages[0]
	}

	if selectedImageURL == "" {
		job.fail("No suitable product image found")
		return
	}

	// Download final image
	job.updateStatus("processing", "Downloading image...")
	client := &http.Client{Timeout: 30 * time.Second}
	finalImgResp, err := client.Get(selectedImageURL)
	if err != nil {
		job.fail("Failed to download image: " + err.Error())
		return
	}
	defer finalImgResp.Body.Close()

	finalImgBytes, err := io.ReadAll(finalImgResp.Body)
	if err != nil {
		job.fail("Failed to read image data")
		return
	}

	// Remove background (if rembg available)
	if rembgAvailable() {
		job.updateStatus("processing", "Removing background...")
		if cleanedBytes, err := removeBackgroundFromBytes(finalImgBytes); err == nil {
			finalImgBytes = cleanedBytes
			log.Printf("✅ Background removed for scraped image")
		} else {
			log.Printf("⚠️ Background removal failed, using original: %v", err)
		}
	}

	contentType := http.DetectContentType(finalImgBytes)
	ext := ".jpg"
	if contentType == "image/png" {
		ext = ".png"
	} else if contentType == "image/webp" {
		ext = ".webp"
	}

	var finalURL string

	// Upload to S3 if available
	if s3Client != nil {
		filename := fmt.Sprintf("items/scraped_%d%s", time.Now().UnixNano(), ext)
		s3URL, err := uploadBytesToS3(filename, finalImgBytes, contentType)
		if err == nil {
			finalURL = s3URL
		} else {
			log.Printf("⚠️ S3 upload failed: %v, falling back to local storage", err)
		}
	}

	// Fallback to local storage if S3 failed or not configured
	if finalURL == "" {
		filename := fmt.Sprintf("scraped_%d%s", time.Now().UnixNano(), ext)
		localDir := "../images"
		os.MkdirAll(localDir, 0755)
		localPath := filepath.Join(localDir, filename)
		if err := os.WriteFile(localPath, finalImgBytes, 0644); err != nil {
			job.fail("Failed to save image locally: " + err.Error())
			return
		}
		finalURL = fmt.Sprintf("http://localhost:%s/images/%s", getEnv("PORT", "8556"), filename)
	}

	log.Printf("✅ Scrape Job Complete: %s -> %s", req.URL, finalURL)

	job.complete(ScrapeResponse{
		Title:       title,
		ImageURL:    finalURL,
		Description: description,
		SiteName:    siteName,
		OriginalURL: req.URL,
		Category:    category,
		Color:       color,
		Brand:       brand,
		Price:       price,
	})
}

// handleSegmentOutfit handles outfit segmentation requests
func handleSegmentOutfit(w http.ResponseWriter, r *http.Request) {
	log.Printf("👗 Outfit segmentation request received")

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the uploaded image
	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error getting uploaded file: %v", err)
		http.Error(w, "No image file provided", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read the image data
	imageData, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading image data: %v", err)
		http.Error(w, "Failed to read image", http.StatusInternalServerError)
		return
	}

	// Encode to base64 for AI processing
	imageBase64 := base64.StdEncoding.EncodeToString(imageData)

	log.Printf("👗 Processing outfit image: %s (%.1fKB)", header.Filename, float64(len(imageData))/1024)

	// Use AI to analyze the outfit and detect individual items
	segments, err := segmentOutfitWithAI(imageBase64)
	if err != nil {
		log.Printf("Error segmenting outfit: %v", err)
		http.Error(w, "Failed to segment outfit: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"segments": segments,
	})
}

// OutfitSegment represents a detected clothing item in the outfit
type OutfitSegment struct {
	Category    string  `json:"category"`
	BoundingBox BBox    `json:"bounding_box"`
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

// BBox represents a bounding box
type BBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// segmentOutfitWithAI uses AI to detect and segment clothing items in an outfit photo
func segmentOutfitWithAI(imageBase64 string) ([]OutfitSegment, error) {
	log.Printf("👗 Using AI to segment outfit...")

	// Use Ollama vision model to analyze the image and detect clothing items
	prompt := `Analyze this outfit image and identify individual clothing items. For each item you can see, provide:
1. Category (tops, bottoms, shoes, outerwear, accessories)
2. A bounding box (approximate x, y, width, height as percentages of image size)
3. Confidence level (0.0 to 1.0)
4. Brief description

Respond in JSON format with an array of detected items. Even if items overlap, try to identify distinct pieces.

Example response:
{
  "items": [
    {
      "category": "tops",
      "bounding_box": {"x": 20, "y": 10, "width": 60, "height": 40},
      "confidence": 0.9,
      "description": "white cotton t-shirt"
    },
    {
      "category": "bottoms", 
      "bounding_box": {"x": 25, "y": 50, "width": 50, "height": 45},
      "confidence": 0.8,
      "description": "blue denim jeans"
    }
  ]
}`

	// Call Ollama
	payload := map[string]interface{}{
		"model":  visionModel,
		"prompt": prompt,
		"stream": false,
		"images": []string{imageBase64},
		"options": map[string]interface{}{
			"temperature": 0.1, // Low temperature for consistent analysis
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(ollamaURL+"/api/generate", "application/json", bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Ollama returned status %d", resp.StatusCode)
	}

	var ollamaResp struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	// Parse the AI response
	var aiResult struct {
		Items []struct {
			Category    string  `json:"category"`
			BoundingBox BBox    `json:"bounding_box"`
			Confidence  float64 `json:"confidence"`
			Description string  `json:"description"`
		} `json:"items"`
	}

	// Try to extract JSON from the response
	response := strings.TrimSpace(ollamaResp.Response)

	// Find JSON content (sometimes wrapped in markdown code blocks)
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonContent := response[jsonStart : jsonEnd+1]
		if err := json.Unmarshal([]byte(jsonContent), &aiResult); err != nil {
			log.Printf("Failed to parse AI response as JSON: %v", err)
			log.Printf("Raw response: %s", response)
			// Return a fallback response
			return []OutfitSegment{
				{
					Category:    "tops",
					BoundingBox: BBox{X: 25, Y: 20, Width: 50, Height: 35},
					Confidence:  0.5,
					Description: "detected top item",
				},
				{
					Category:    "bottoms",
					BoundingBox: BBox{X: 25, Y: 55, Width: 50, Height: 40},
					Confidence:  0.5,
					Description: "detected bottom item",
				},
			}, nil
		}
	} else {
		log.Printf("No valid JSON found in AI response: %s", response)
		return []OutfitSegment{}, fmt.Errorf("AI response did not contain valid JSON")
	}

	// Convert to our segment format
	var segments []OutfitSegment
	for _, item := range aiResult.Items {
		segments = append(segments, OutfitSegment{
			Category:    item.Category,
			BoundingBox: item.BoundingBox,
			Confidence:  item.Confidence,
			Description: item.Description,
		})
	}

	log.Printf("👗 Detected %d clothing items in outfit", len(segments))
	return segments, nil
}

// ============================================================================
// ANALYTICS ENDPOINTS
// ============================================================================

// handleWardrobeStats returns overall wardrobe statistics
func handleWardrobeStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := int64(1) // Default user for PoC

	stats, err := toolGetWardrobeStats(ctx, userID)
	if err != nil {
		log.Printf("Database query failed, returning empty stats: %v", err)
		// Return empty stats structure for PoC
		stats = map[string]interface{}{
			"total_items":   0,
			"total_outfits": 0,
			"by_category":   map[string]interface{}{},
			"by_color":      map[string]interface{}{},
			"by_season":     map[string]interface{}{},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleMostWorn returns most frequently worn items
func handleMostWorn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := int64(1) // Default user for PoC

	items, err := toolGetMostWorn(ctx, userID, 10) // Top 10 most worn items
	if err != nil {
		log.Printf("Database query failed, returning empty items: %v", err)
		// Return empty array for PoC
		items = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// handleLeastWorn returns least frequently worn items
func handleLeastWorn(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := int64(1) // Default user for PoC

	items, err := toolGetLeastWorn(ctx, userID, 10) // Top 10 least worn items
	if err != nil {
		log.Printf("Database query failed, returning empty items: %v", err)
		// Return empty array for PoC
		items = []map[string]interface{}{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}
