package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	appConfig   *Config
	configMutex sync.RWMutex
	configPath  string
)

type Config struct {
	Server          ServerConfig          `yaml:"server" json:"server"`
	AI              AIConfig              `yaml:"ai" json:"ai"`
	ImageProcessing ImageProcessingConfig `yaml:"image_processing" json:"imageProcessing"`
	Storage         StorageConfig         `yaml:"storage" json:"storage"`
	Scraper         ScraperConfig         `yaml:"scraper" json:"scraper"`
	Weather         WeatherConfig         `yaml:"weather" json:"weather"`
	Logging         LoggingConfig         `yaml:"logging" json:"logging"`
}

type ServerConfig struct {
	Port int    `yaml:"port" json:"port"`
	Host string `yaml:"host" json:"host"`
}

type AIConfig struct {
	Provider string       `yaml:"provider" json:"provider"` // "local" or "cloud"
	Ollama   OllamaConfig `yaml:"ollama" json:"ollama"`
	Cloud    CloudConfig  `yaml:"cloud" json:"cloud"`
}

type OllamaConfig struct {
	URL          string `yaml:"url" json:"url"`
	VisionModel  string `yaml:"vision_model" json:"visionModel"`
	TextModel    string `yaml:"text_model" json:"textModel"`
	SpatialModel string `yaml:"spatial_model" json:"spatialModel"`
}

type CloudConfig struct {
	OpenAIAPIKey    string `yaml:"openai_api_key" json:"openaiApiKey"`
	OpenAIModel     string `yaml:"openai_model" json:"openaiModel"`
	AnthropicAPIKey string `yaml:"anthropic_api_key" json:"anthropicApiKey"`
	AnthropicModel  string `yaml:"anthropic_model" json:"anthropicModel"`
}

type ImageProcessingConfig struct {
	RembgURL      string `yaml:"rembg_url" json:"rembgUrl"`
	AutoRemoveBg  bool   `yaml:"auto_remove_bg" json:"autoRemoveBg"`
	VTONBackend   string `yaml:"vton_backend" json:"vtonBackend"`
	CatVTONURL    string `yaml:"catvton_url" json:"catvtonUrl"`
	LadiVTONShoes string `yaml:"ladivton_shoes_url" json:"ladivtonShoesUrl"`
}

type StorageConfig struct {
	S3 S3Config `yaml:"s3" json:"s3"`
}

type S3Config struct {
	Enabled   bool   `yaml:"enabled" json:"enabled"`
	Endpoint  string `yaml:"endpoint" json:"endpoint"`
	Bucket    string `yaml:"bucket" json:"bucket"`
	AccessKey string `yaml:"access_key" json:"accessKey"`
	SecretKey string `yaml:"secret_key" json:"secretKey"`
	Region    string `yaml:"region" json:"region"`
}

type ScraperConfig struct {
	UserAgent        string `yaml:"user_agent" json:"userAgent"`
	TimeoutSeconds   int    `yaml:"timeout_seconds" json:"timeoutSeconds"`
	AIImageSelection bool   `yaml:"ai_image_selection" json:"aiImageSelection"`
	AutoRemoveBg     bool   `yaml:"auto_remove_bg" json:"autoRemoveBg"`
}

type WeatherConfig struct {
	Location string `yaml:"location" json:"location"`
	Unit     string `yaml:"unit" json:"unit"` // "celsius" or "fahrenheit"
}

type LoggingConfig struct {
	Level  string `yaml:"level" json:"level"`
	Format string `yaml:"format" json:"format"`
}

func getDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: 8556,
			Host: "0.0.0.0",
		},
		AI: AIConfig{
			Provider: "local",
			Ollama: OllamaConfig{
				URL:          "http://localhost:11434",
				VisionModel:  "llava:34b",
				TextModel:    "llama3.2:3b",
				SpatialModel: "qwen3-vl:32b",
			},
			Cloud: CloudConfig{
				OpenAIModel:    "gpt-4-vision-preview",
				AnthropicModel: "claude-3-opus-20240229",
			},
		},
		ImageProcessing: ImageProcessingConfig{
			RembgURL:      "http://localhost:5000",
			AutoRemoveBg:  true,
			VTONBackend:   "fashn",
			CatVTONURL:    "http://localhost:7860",
			LadiVTONShoes: "http://localhost:8558",
		},
		Storage: StorageConfig{
			S3: S3Config{
				Enabled:   true,
				Endpoint:  "http://localhost:8333",
				Bucket:    "gothreads",
				AccessKey: "admin",
				SecretKey: "admin123",
				Region:    "us-east-1",
			},
		},
		Scraper: ScraperConfig{
			UserAgent:        "Mozilla/5.0 (compatible; GoThreads/1.0)",
			TimeoutSeconds:   30,
			AIImageSelection: true,
			AutoRemoveBg:     true,
		},
		Weather: WeatherConfig{
			Location: "London",
			Unit:     "celsius",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

func loadConfig() (*Config, error) {
	// Try multiple config paths
	configPaths := []string{
		"config.yaml",
		"../config.yaml",
		"./config.yaml",
		"/etc/gothreads/config.yaml",
	}

	// Allow override via env
	if envPath := os.Getenv("GOTHREADS_CONFIG"); envPath != "" {
		configPaths = append([]string{envPath}, configPaths...)
	}

	for _, p := range configPaths {
		if _, err := os.Stat(p); err == nil {
			configPath = p
			break
		}
	}

	if configPath == "" {
		log.Println("No config.yaml found, using defaults + env vars")
		return applyEnvOverrides(getDefaultConfig()), nil
	}

	log.Printf("Loading config from: %s", configPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	config := getDefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Apply env overrides
	config = applyEnvOverrides(config)

	return config, nil
}

func applyEnvOverrides(config *Config) *Config {
	// Server
	if v := os.Getenv("PORT"); v != "" {
		fmt.Sscanf(v, "%d", &config.Server.Port)
	}

	// AI
	if v := os.Getenv("AI_PROVIDER"); v != "" {
		config.AI.Provider = v
	}
	if v := os.Getenv("OLLAMA_URL"); v != "" {
		config.AI.Ollama.URL = v
	}
	if v := os.Getenv("OLLAMA_VISION_MODEL"); v != "" {
		config.AI.Ollama.VisionModel = v
	}
	if v := os.Getenv("OLLAMA_TEXT_MODEL"); v != "" {
		config.AI.Ollama.TextModel = v
	}
	if v := os.Getenv("OLLAMA_SPATIAL_MODEL"); v != "" {
		config.AI.Ollama.SpatialModel = v
	}
	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		config.AI.Cloud.OpenAIAPIKey = v
	}
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		config.AI.Cloud.AnthropicAPIKey = v
	}

	// Image Processing
	if v := os.Getenv("REMBG_URL"); v != "" {
		config.ImageProcessing.RembgURL = v
	}
	if v := os.Getenv("VTON_BACKEND_URL"); v != "" {
		config.ImageProcessing.VTONBackend = v
	}
	if v := os.Getenv("CATVTON_LOCAL_URL"); v != "" {
		config.ImageProcessing.CatVTONURL = v
	}
	if v := os.Getenv("LADIVTON_SHOES_URL"); v != "" {
		config.ImageProcessing.LadiVTONShoes = v
	}

	// Storage
	if v := os.Getenv("S3_ENDPOINT"); v != "" {
		config.Storage.S3.Endpoint = v
	}
	if v := os.Getenv("S3_BUCKET"); v != "" {
		config.Storage.S3.Bucket = v
	}
	if v := os.Getenv("S3_ACCESS_KEY"); v != "" {
		config.Storage.S3.AccessKey = v
	}
	if v := os.Getenv("S3_SECRET_KEY"); v != "" {
		config.Storage.S3.SecretKey = v
	}

	return config
}

func saveConfig(config *Config) error {
	if configPath == "" {
		// Default to config.yaml in current directory
		configPath = "config.yaml"
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Create directory if needed
	dir := filepath.Dir(configPath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	log.Printf("Config saved to: %s", configPath)
	return nil
}

func getConfig() *Config {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return appConfig
}

func updateConfig(newConfig *Config) {
	configMutex.Lock()
	defer configMutex.Unlock()
	appConfig = newConfig

	// Update global variables that other parts of the code use
	ollamaURL = newConfig.AI.Ollama.URL
	visionModel = newConfig.AI.Ollama.VisionModel
	textModel = newConfig.AI.Ollama.TextModel
	spatialModel = newConfig.AI.Ollama.SpatialModel
	rembgURL = newConfig.ImageProcessing.RembgURL
	vtonBackendURL = newConfig.ImageProcessing.VTONBackend
	catvtonLocalURL = newConfig.ImageProcessing.CatVTONURL
	ladivtonShoesURL = newConfig.ImageProcessing.LadiVTONShoes
}

// handleGetConfig returns the current configuration (without sensitive data)
func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	config := getConfig()

	// Create a safe copy without sensitive data
	safeConfig := *config
	safeConfig.AI.Cloud.OpenAIAPIKey = maskAPIKey(config.AI.Cloud.OpenAIAPIKey)
	safeConfig.AI.Cloud.AnthropicAPIKey = maskAPIKey(config.AI.Cloud.AnthropicAPIKey)
	safeConfig.Storage.S3.SecretKey = maskAPIKey(config.Storage.S3.SecretKey)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(safeConfig)
}

// handleUpdateConfig updates the configuration
func handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var updates Config
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}

	currentConfig := getConfig()

	// Merge updates (only non-empty fields)
	if updates.AI.Provider != "" {
		currentConfig.AI.Provider = updates.AI.Provider
	}
	if updates.AI.Ollama.URL != "" {
		currentConfig.AI.Ollama.URL = updates.AI.Ollama.URL
	}
	if updates.AI.Ollama.VisionModel != "" {
		currentConfig.AI.Ollama.VisionModel = updates.AI.Ollama.VisionModel
	}
	if updates.AI.Ollama.TextModel != "" {
		currentConfig.AI.Ollama.TextModel = updates.AI.Ollama.TextModel
	}
	if updates.AI.Ollama.SpatialModel != "" {
		currentConfig.AI.Ollama.SpatialModel = updates.AI.Ollama.SpatialModel
	}
	// Only update API keys if they're not masked
	if updates.AI.Cloud.OpenAIAPIKey != "" && !isMasked(updates.AI.Cloud.OpenAIAPIKey) {
		currentConfig.AI.Cloud.OpenAIAPIKey = updates.AI.Cloud.OpenAIAPIKey
	}
	if updates.AI.Cloud.AnthropicAPIKey != "" && !isMasked(updates.AI.Cloud.AnthropicAPIKey) {
		currentConfig.AI.Cloud.AnthropicAPIKey = updates.AI.Cloud.AnthropicAPIKey
	}

	if updates.ImageProcessing.RembgURL != "" {
		currentConfig.ImageProcessing.RembgURL = updates.ImageProcessing.RembgURL
	}
	if updates.ImageProcessing.VTONBackend != "" {
		currentConfig.ImageProcessing.VTONBackend = updates.ImageProcessing.VTONBackend
	}
	currentConfig.ImageProcessing.AutoRemoveBg = updates.ImageProcessing.AutoRemoveBg

	if updates.Weather.Location != "" {
		currentConfig.Weather.Location = updates.Weather.Location
	}
	if updates.Weather.Unit != "" {
		currentConfig.Weather.Unit = updates.Weather.Unit
	}

	if err := saveConfig(currentConfig); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save config: "+err.Error())
		return
	}

	updateConfig(currentConfig)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func isMasked(key string) bool {
	return len(key) > 8 && key[4:8] == "****"
}
