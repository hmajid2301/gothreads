# Mockup Server with AI

Throwaway Go service that serves the HTML mockups and provides AI features via Ollama.

## Requirements

- Go 1.21+
- Ollama running locally with `llava:7b` model

## Quick Start

```bash
# Terminal 1: Start Ollama (skip if using Docker)
ollama serve

# Terminal 2: Pull the vision model (first time only)
ollama pull llava:7b
ollama pull llama3.2:3b  # optional, for text generation

# Terminal 3: Start the mockup server
task mockups:dev
# or manually:
cd mockups/api
go run main.go
```

The server runs on `http://localhost:8556` by default and serves both the static mockup files and the AI API.

## Endpoints

### Static Files
- `GET /` - Serves mockup HTML/CSS/JS files from parent directory
- `GET /upload.html` - Upload page with AI integration

### API Endpoints (all prefixed with `/api`)

#### GET /api/health
Health check.

#### GET /api/status
Check Ollama connection and available models.

```bash
curl http://localhost:8556/api/status
```

#### POST /api/analyze
Analyze a clothing image with the vision model.

```bash
curl -X POST http://localhost:8556/api/analyze \
  -H "Content-Type: application/json" \
  -d '{"image": "data:image/jpeg;base64,..."}'
```

Response:
```json
{
  "description": "Navy blue cotton t-shirt with crew neck",
  "category": "Tops",
  "color": "Navy",
  "tags": ["casual", "cotton", "summer", "basic"]
}
```

#### POST /api/tags
Generate tags from a description (uses text model).

```bash
curl -X POST http://localhost:8556/api/tags \
  -H "Content-Type: application/json" \
  -d '{"description": "Blue denim jeans, slim fit"}'
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8556 | Server port |
| `OLLAMA_URL` | http://localhost:11434 | Ollama API URL |
| `OLLAMA_VISION_MODEL` | llava:7b | Model for image analysis |
| `OLLAMA_TEXT_MODEL` | llama3.2:3b | Model for text generation |

## Using with Docker Ollama

If Ollama is running in Docker:

```bash
OLLAMA_URL=http://localhost:11434 go run main.go
```

## Notes

This is PoC code for the mockups. The real implementation will be in the main backend.
