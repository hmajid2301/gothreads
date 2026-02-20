# External Integrations

**Analysis Date:** 2026-02-20

## APIs & External Services

**AI/ML Services:**
- Ollama - Local AI inference server
  - SDK/Client: Direct HTTP calls
  - Configuration: `config.yaml` ollama section
  - Models: llava:34b (vision), llama3.2:3b (text), qwen3-vl:32b (spatial)

**Cloud AI (Optional):**
- OpenAI - GPT-4 Vision API
  - Auth: OPENAI_API_KEY environment variable
- Anthropic Claude - Vision and text processing
  - Auth: ANTHROPIC_API_KEY environment variable

**Image Processing:**
- rembg - Background removal service
  - Endpoint: http://localhost:5000
  - Docker: danielgatis/rembg:latest

**Virtual Try-On:**
- CATVTon - Local virtual try-on service
  - Endpoint: http://localhost:7860
  - Container: Custom ROCm/PyTorch build

## Data Storage

**Databases:**
- PostgreSQL 18
  - Connection: GOTHREADS_DB_DATABASE_URL
  - Client: pgx/v5 with connection pooling
  - Migrations: Goose migration tool

**File Storage:**
- SeaweedFS - S3-compatible object storage
  - Endpoint: http://localhost:8333
  - Bucket: gothreads
  - Access credentials in `config.yaml`

**Caching:**
- None (direct database access)

## Authentication & Identity

**Auth Provider:**
- OAuth 2.0 multi-provider support
  - Implementation: Custom JWT-based sessions
  - Providers: Authentik, Authelia, Google, GitHub
  - Session storage: HTTP-only cookies

**Configuration:**
- Environment variables per provider (OAUTH_PROVIDER_CLIENT_ID, etc.)
- JWT signing: JWT_SECRET environment variable
- Skip auth mode: SKIP_AUTH=true for development

## Monitoring & Observability

**Error Tracking:**
- Standard Go log package

**Logs:**
- Console output
- Configurable level and format in `config.yaml`

## CI/CD & Deployment

**Hosting:**
- Docker Compose for development
- Nix builds for production containers

**CI Pipeline:**
- None configured (local builds via Nix)

## Environment Configuration

**Required env vars:**
- GOTHREADS_DB_DATABASE_URL - PostgreSQL connection string
- JWT_SECRET - JWT token signing key
- OAuth provider credentials (optional)

**Secrets location:**
- Environment variables
- `.env` files (noted but not read for security)

## Webhooks & Callbacks

**Incoming:**
- OAuth callback endpoints (`/auth/{provider}/callback`)

**Outgoing:**
- None configured

## Development Services

**Docker Compose Services:**
- postgres:18 - Primary database (port 15433)
- seaweedfs - S3-compatible storage (ports 8333, 9333, 19333)
- rembg - Background removal API (port 5000)
- catvton - Virtual try-on service (port 7860)
- migrate - Database migration runner (gomicro/goose:3.25.0)

**Service Dependencies:**
- CATVTon requires GPU with ROCm support
- HuggingFace model caching via volumes

---

*Integration audit: 2026-02-20*