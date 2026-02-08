# go-threads Mockup Index

Quick navigation for all mockup pages.

## 📱 View Mockups

**Option 1: Direct Open**
```bash
cd mockups
open index.html  # macOS
xdg-open index.html  # Linux
start index.html  # Windows
```

**Option 2: Local Server**
```bash
cd mockups
python3 -m http.server 8000
# Visit: http://localhost:8000
```

---

## 🗂️ Page Navigation

### Core Flow
1. **[Home](index.html)** - Dashboard with stats and quick actions
2. **[Upload](upload.html)** - Drag-drop file upload
3. **[Item Detail](item-detail.html)** - Edit metadata after upload
4. **[Wardrobe](wardrobe.html)** - Grid view of all items

### Outfit Creation (Main Feature)
5. **[Outfit Builder](outfit-builder.html)** ⭐ - Drag-drop canvas for creating outfits
6. **[Outfits Gallery](outfits.html)** - View saved outfits

### Insights & Settings
7. **[Analytics](analytics.html)** - Charts, stats, most/least worn items
8. **[Settings](settings.html)** - Preferences, AI config, data export

---

## 🎯 Key Features Demonstrated

### Upload Flow
```
Upload Page → (Select Image) → Item Detail Page → (Edit Metadata) → Wardrobe
```

### Outfit Creation Flow
```
Outfit Builder → (Click items to add) → (Drag to position) → (Save) → Outfits Gallery
```

---

## 🎨 Design Elements

**Color Scheme** (Catppuccin-inspired):
- Primary: `#FFBCBA` (Warm Pink)
- Background: `#FFF4E9` (Cream)
- Text: `#553630` (Dark Brown)

**Key UI Patterns**:
- Cards with soft shadows
- Rounded corners (0.75rem)
- Responsive grid layouts
- Drag-and-drop interactions
- Toast notifications

---

## 🔍 What to Review

### Critical Interactions
- [ ] **Upload experience** - Is drag-drop intuitive?
- [ ] **Outfit canvas** - Does drag positioning make sense?
- [ ] **Item picker** - Easy to find and add items?
- [ ] **Navigation** - Clear page hierarchy?

### Visual Design
- [ ] **Color palette** - Warm and inviting?
- [ ] **Typography** - Readable and hierarchical?
- [ ] **Spacing** - Comfortable and balanced?
- [ ] **Mobile responsiveness** - Works on small screens?

### User Flows
- [ ] **Upload to wardrobe** - Smooth and obvious?
- [ ] **Create outfit** - Clear steps?
- [ ] **View analytics** - Meaningful insights?
- [ ] **Settings management** - Easy to configure?

---

## 📝 Feedback

When reviewing, consider:

1. **What feels confusing?**
2. **What's missing?**
3. **What would you change?**
4. **What do you love?**

---

## ⚠️ Remember

These are **throwaway prototypes**! They will be completely rebuilt with:
- HTMX + Templ for server-side rendering
- Alpine.js for simple interactions
- Svelte component for outfit canvas
- Real backend with PostgreSQL

The goal is to validate UX flows, not build production code.
