# go-threads Mockups

**⚠️ Throwaway Prototypes** - These HTML/CSS/JS mockups are for visualization only and will be replaced with HTMX + Templ + Alpine.js in the real application.

## Purpose

Visual proof-of-concept for:
1. **Upload flow** - How users add items to their wardrobe
2. **Outfit builder** - Drag-and-drop canvas for creating outfits

## Pages

### 1. `index.html` - Home Dashboard
- Quick stats overview (total items, outfits, cost-per-wear)
- Recent items grid
- Quick action buttons

### 2. `upload.html` - Upload New Item
- Drag-and-drop file upload area
- File input with validation hints
- Tips for great photos
- After upload → redirects to item detail page

### 3. `item-detail.html` - Edit Item Details
- Image preview (with AI background removal indicator)
- Form for item metadata (name, category, price, brand, color, season, notes)
- Save/Cancel actions

### 4. `wardrobe.html` - Wardrobe Grid
- Grid view of all items (4 columns)
- Category filter dropdown
- Add item button
- Click item to view details

### 5. `outfit-builder.html` - Outfit Canvas ⭐ (MOST IMPORTANT)
- **Left**: Drag-and-drop canvas for positioning items
- **Right**: Item picker sidebar
- Click items to add to canvas
- Drag items to position and arrange
- Clear canvas button
- Save outfit button

### 6. `outfits.html` - Outfit Gallery
- Grid view of saved outfits (3 columns)
- Each outfit shows composite preview of items
- Wear count and rating display
- Create new outfit card (dashed border)
- Tips section

### 7. `analytics.html` - Analytics Dashboard
- Stats overview (4 cards)
- Category breakdown (pie chart placeholder)
- Wear frequency chart (line chart placeholder)
- Most worn items table
- Least worn items table
- Spending over time (bar chart placeholder)

### 8. `settings.html` - Settings & Preferences
- Profile settings (name, email)
- AI processing preferences (background removal, auto-tagging, provider selection)
- Notifications (push, email summaries)
- Wardrobe preferences (max wears, currency, date format)
- Data & privacy (export, delete account)
- About section (version, system status)

## How to View

1. Open any HTML file in a web browser
2. No build step required - pure HTML/CSS/JS

```bash
# From mockups directory
open index.html
# or
python3 -m http.server 8000
# Then visit http://localhost:8000
```

## Key Features Demonstrated

### Upload Flow
1. User uploads image via drag-drop or file picker
2. Image preview shown immediately
3. AI background removal processing indicator
4. Form to add metadata (name, category, price, etc.)
5. Save to wardrobe

### Outfit Builder (Core Feature)
1. Canvas area for visual composition
2. Item picker sidebar with thumbnails
3. Click item → adds to canvas at random position
4. Drag items to arrange
5. Items can overlap (z-index layering)
6. Save outfit with name

## Design System

**Color Palette** (Catppuccin-inspired):
- Primary: `#FFBCBA` (Warm Pink)
- Background: `#FFF4E9` (Warm Cream)
- Text: `#553630` (Dark Brown)

**Typography**:
- System font stack (no web fonts for fast loading)

**Components**:
- Cards with rounded corners
- Soft shadows
- Responsive grid layouts

## Differences from Real App

| Mockup | Real Implementation |
|--------|---------------------|
| Vanilla JS drag-drop | Svelte component with proper state |
| Local file upload | Multipart POST to Go backend |
| Mock data in JS | PostgreSQL + API endpoints |
| No persistence | Data saved to database |
| Client-side routing | Server-side HTMX navigation |
| No auth | OAuth2/OIDC integration |
| No AI processing | Real rembg background removal |

## Next Steps

After approval of mockup design:
1. ✅ Copy color scheme to TailwindCSS config
2. ✅ Adapt layouts to Templ templates
3. ✅ Build Svelte OutfitCanvas component
4. ✅ Implement backend endpoints
5. ❌ Delete this mockups directory

## Notes

- **File size**: All pages < 50KB combined (excluding placeholder images)
- **Browser compat**: Works in all modern browsers
- **Responsive**: Basic responsive breakpoints included
- **Accessibility**: Semantic HTML, but not fully WCAG compliant (will fix in real app)

---

**Status**: 🎨 Mockup Phase - Ready for feedback
