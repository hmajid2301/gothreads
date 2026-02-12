# AI Features Documentation

## Overview

gothreads uses AI to enhance your digital wardrobe experience. All AI features run locally via Ollama for privacy and speed.

## 🎙️ Voice Input Support

**NEW: Talk to Your Wardrobe**

You can now use voice notes to interact with your wardrobe naturally, without typing.

### Supported Actions via Voice:

1. **Add Items**
   - "Add a blue denim jacket, price $89"
   - "Upload my new red sneakers"
   - "I bought a white t-shirt for $25"

2. **Create Outfits**
   - "Create a casual outfit with my blue jeans and white shirt"
   - "Make an outfit for a business meeting"
   - "I need something for a date night"

3. **Search & Browse**
   - "Show me all my blue items"
   - "What tops do I have?"
   - "Find my least worn clothes"

4. **Get Suggestions**
   - "What should I wear tomorrow?"
   - "Suggest an outfit for sunny weather"
   - "What goes well with my black pants?"

5. **Analytics Queries**
   - "How often do I wear my red jacket?"
   - "What's my most expensive item?"
   - "Show my cost per wear stats"

### How to Use Voice Input:

1. **Click the microphone button** 🎤 (available on chat interface and upload page)
2. **Speak your request** clearly
3. **AI transcribes and processes** your voice note
4. **Get results** instantly

### Technical Details:

- **Transcription**: Uses Web Speech API (browser-based) or Whisper model (Ollama)
- **Language Support**: English (primary), expandable to other languages
- **Privacy**: All processing happens locally - no cloud services
- **Accuracy**: ~95% transcription accuracy for clear speech
- **Fallback**: If voice fails, you can still type

### Voice Commands Reference:

| Command Type | Example Voice Input |
|-------------|---------------------|
| Upload | "I want to add a new item" |
| Search | "Find my blue jeans" |
| Outfit | "Create a workout outfit" |
| Analytics | "Show me my wardrobe stats" |
| Calendar | "What should I wear on Friday?" |
| Filter | "Show only my tops" |

## 🤖 Core AI Features

### 1. Automatic Background Removal

**What it does**: Removes backgrounds from uploaded clothing photos automatically

**How it works**:
- Uses rembg (u2net model) service
- Runs on every upload
- Returns PNG with transparent background

**Endpoint**: `POST /api/remove-bg`

**Settings**: Can be disabled in Settings > AI Processing

### 2. Clothing Analysis & Metadata Detection

**What it does**: Automatically detects category, color, brand, and tags

**How it works**:
- Uses llava:7b vision model
- Analyzes clothing item images
- Extracts metadata in JSON format

**Endpoint**: `POST /api/analyze`

**Request**:
```json
{
  "image": "data:image/png;base64,...",
  "model": "llava:7b"  // optional
}
```

**Response**:
```json
{
  "description": "A beige cable-knit sweater",
  "category": "Tops",
  "color": "Beige",
  "brand": "MR MARVIS",
  "tags": ["casual", "knitwear", "warm"]
}
```

### 3. Enhanced Brand Detection

**What it does**: Uses larger model for better brand label reading

**How it works**:
- Uses llava:13b (enhanced) model
- Slower but more accurate
- Reads small text on labels/tags

**Endpoint**: `POST /api/analyze` with `model: "llava:13b"`

**Use case**: When llava:7b fails to detect brand

### 4. AI Outfit Positioning ("Dress with AI")

**What it does**: Automatically positions clothing items on a mannequin/body

**How it works**:
- Analyzes item categories (tops, bottoms, shoes, outerwear)
- Calculates appropriate positions
- Handles proper layering (z-index)

**Endpoint**: `POST /api/dress`

**Request**:
```json
{
  "body_image": "data:image/png;base64,...",
  "items": [
    {
      "id": 1,
      "image": "data:image/png;base64,...",
      "category": "Tops"
    }
  ]
}
```

**Response**:
```json
{
  "positions": [
    {
      "id": 1,
      "x": 100,
      "y": 50,
      "z": 1,
      "scale": 1.0
    }
  ],
  "explanation": "Positioned top on upper body..."
}
```

### 5. Outfit Rating

**What it does**: AI rates your outfit and provides feedback

**Endpoint**: `POST /api/rate-outfit`

**Response**:
```json
{
  "rating": 8.5,
  "stars": 4,
  "feedback": "Great color coordination...",
  "strengths": ["Well balanced", "Good layering"],
  "improvements": ["Try different shoes"]
}
```

### 6. AI Outfit Suggestions

**What it does**: Suggests complete outfits based on occasion/weather

**Endpoint**: `POST /api/suggest-outfit`

**Request**:
```json
{
  "occasion": "casual",
  "weather": "mild",
  "items": [/* wardrobe items */]
}
```

**Response**:
```json
{
  "selected_items": [1, 3, 7],
  "reasoning": "These items work well together...",
  "tips": ["Roll up the sleeves", "Add a watch"]
}
```

### 7. Color Coordination Analysis

**What it does**: Analyzes if selected items' colors work together

**Endpoint**: `POST /api/analyze-colors`

**Response**:
```json
{
  "score": 85,
  "analysis": "Blue and white create a fresh look",
  "suggestions": ["Add brown accessories"]
}
```

### 8. Style Matching

**What it does**: Finds items that match well with a selected piece

**Endpoint**: `POST /api/match-items`

**Response**:
```json
{
  "matches": [
    {
      "item_id": 5,
      "score": 92,
      "reason": "Complementary colors and style"
    }
  ]
}
```

### 9. Wardrobe Gap Analysis

**What it does**: Identifies missing essential items

**Endpoint**: `POST /api/analyze-wardrobe`

**Response**:
```json
{
  "assessment": "Well-rounded wardrobe with 15 items",
  "missing": ["Dark jeans", "White sneakers"],
  "strengths": ["Good variety of tops"],
  "tips": ["Consider adding neutral basics"]
}
```

### 10. AI Analytics Insights

**What it does**: Analyzes wearing patterns and provides insights

**Endpoint**: `POST /api/analytics-insights`

**Response**:
```json
{
  "insights": [
    "You wear blue items 3x more than red",
    "Your jacket is underutilized (worn 2x in 6 months)"
  ],
  "recommendations": ["Wear item #5 more often"]
}
```

### 11. AI Calendar Suggestions

**What it does**: Suggests outfits for calendar events

**Endpoint**: `POST /api/calendar-suggestions`

**Request**:
```json
{
  "date": "2026-02-15",
  "event": "Team meeting",
  "weather": "cold"
}
```

### 12. Outfit Regeneration/Variations

**What it does**: Creates variations of existing outfits

**Endpoint**: `POST /api/outfit-regenerate`

**Request**:
```json
{
  "outfit_id": 5,
  "modifier": "more formal"  // or "more casual"
}
```

### 13. Chat Interface with Voice Support

**What it does**: Natural language interaction with your wardrobe using text or voice

**Endpoint**: `POST /api/chat-outfit`

**Request**:
```json
{
  "message": "Show me a casual look",
  "context": {
    "items": [/* user's wardrobe */],
    "recent_outfits": [/* recent creations */]
  }
}
```

**Response**:
```json
{
  "reply": "I'd suggest pairing your blue jeans with...",
  "suggested_items": [1, 3, 7],
  "action": "create_outfit"
}
```

**Voice Input**: Click microphone to speak instead of typing

## 🎯 Performance Benchmarks

**CPU-only performance** (typical):
- Background removal: 5-10 seconds
- Image analysis (llava:7b): 60-90 seconds
- Enhanced brand (llava:13b): 2-3 minutes
- Outfit positioning: 60-90 seconds
- Outfit suggestions: 10-20 seconds
- Color analysis: 5-15 seconds
- Style matching: 10-20 seconds
- Wardrobe gaps: 15-30 seconds
- Outfit rating: 10-20 seconds
- Voice transcription: 1-3 seconds

**GPU performance** (if available):
- 5-10x faster for vision models
- Background removal: <1 second
- Image analysis: 5-10 seconds

## 🔧 Configuration

### AI Models Required

Install via Ollama:
```bash
ollama pull llava:7b      # 4.7GB - Primary vision model
ollama pull llava:13b     # 8.0GB - Enhanced brand detection
ollama pull llama3.2:3b   # 2.0GB - Text generation
ollama pull whisper:tiny  # Optional - Voice transcription
```

### Settings

Configure in Settings page:
- **Auto-remove backgrounds**: On/Off
- **Auto-tag items**: On/Off
- **Preferred AI provider**: Local (Ollama) / Cloud (OpenAI, Anthropic)
- **Voice input**: Enable/Disable
- **Transcription model**: Web Speech API / Whisper

### Environment Variables

```bash
OLLAMA_URL=http://localhost:11434
OLLAMA_VISION_MODEL=llava:7b
OLLAMA_TEXT_MODEL=llama3.2:3b
REMBG_URL=http://localhost:5000
WHISPER_MODEL=whisper:tiny  # For voice transcription
```

## 🐛 Troubleshooting

### AI Not Available
```bash
# Check Ollama status
ollama list

# Restart Ollama
systemctl restart ollama

# Pull missing models
ollama pull llava:7b
```

### Background Removal Fails
```bash
# Check rembg service
docker ps | grep rembg
curl http://localhost:5000

# Restart service
docker compose restart rembg
```

### Voice Input Not Working
```bash
# Check browser support
# Chrome/Edge: Full support
# Firefox: Limited support
# Safari: Requires permissions

# Check microphone permissions
# Settings > Privacy > Microphone

# Fallback to typing if voice fails
```

### Slow Performance
- Expected on CPU-only systems
- Use llava:7b instead of llava:13b
- Consider GPU acceleration
- Reduce image sizes before upload

### Brand Detection Fails
- Use "✨ Detect with AI (13b)" button for enhanced accuracy
- Ensure label/tag is visible in photo
- Good lighting improves detection
- Try retry if first attempt fails

## 🔒 Privacy

- ✅ All AI processing runs locally (Ollama)
- ✅ No data sent to cloud services
- ✅ Images stay on your device
- ✅ Voice transcription can use browser (no server)
- ✅ Optional cloud providers (disabled by default)

## 📊 API Response Format

All AI endpoints follow this format:

**Success**:
```json
{
  "status": "success",
  "data": {/* endpoint-specific data */},
  "processing_time": "2.3s"
}
```

**Error**:
```json
{
  "status": "error",
  "error": "Model not available",
  "code": 503
}
```

## 🚀 Future Features

- [ ] Multi-language voice support
- [ ] Style learning from user preferences
- [ ] Seasonal outfit suggestions
- [ ] Shopping recommendations
- [ ] Virtual try-on
- [ ] Outfit sharing with AI descriptions
- [ ] Voice-controlled wardrobe browsing

## 📝 Notes

- AI suggestions improve over time with more data
- Brand detection works best with clear, well-lit photos
- Voice input requires microphone permissions
- Outfit positioning works best with transparent PNG images
- All features are optional and can be disabled

## 🆘 Support

For issues or questions:
- Check logs: `docker logs gothreads-api`
- Verify AI status: http://localhost:8556/api/status
- Report bugs: https://github.com/yourusername/gothreads/issues
