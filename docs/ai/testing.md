
  # Start all services
  docker compose up -d

  # Verify Ollama models are installed
  ollama list
  # Should show: llava:7b, llava:13b, llama3.2:3b

  # If missing, pull them:
  ollama pull llava:7b
  ollama pull llava:13b
  ollama pull llama3.2:3b

  # Start mockup server
  task mockups:dev

  # Verify server is running
  curl http://localhost:8556/api/health

  ---
  TEST 1: Multi-Image Upload with Background Removal

  Page: http://localhost:8556/upload.html

  1. Clear existing data (optional):
    - Click "🗑️ Clear All Data" button
    - Confirm deletion
  2. Upload multiple clothing items:
    - Click upload area
    - Select 3-5 clothing photos (JPEG/PNG/WebP/AVIF)
    - Expected: Progress modal shows "Processing 1 of 5
  items..."
    - Expected: Each image:
        - Background automatically removed (transparent PNG)
      - AI analyzes: category, color, brand detection (60-90s
   per image)
  3. Verify results:
    - Redirects to wardrobe with ?recent=true
    - Banner shows: "Recently uploaded items shown first"
    - Items display AI-detected info: ✨ Color • Brand
    - All images have transparent backgrounds

  Success Criteria:
  - ✅ All items uploaded successfully
  - ✅ Backgrounds removed
  - ✅ AI detected category and color for each
  - ✅ Items visible in wardrobe with AI info

  ---
  TEST 2: Enhanced Brand Detection

  Page: http://localhost:8556/item-detail.html?id=1

  1. Navigate to item detail:
    - Click any uploaded item from wardrobe
  2. Test enhanced brand detection:
    - Click "✨ Detect with AI (13b)" button next to Brand
  field
    - Expected: Button shows "⏳ Analyzing (2-3 min)..."
    - Expected: After 2-3 minutes, brand field auto-fills
    - Expected: Toast shows "Brand detected: [BRAND_NAME]"
  3. Save item:
    - Click "Save Item"
    - Verify brand persists

  Success Criteria:
  - ✅ llava:13b model used (check server logs)
  - ✅ Brand detected from clothing label
  - ✅ Brand saved successfully

  ---
  TEST 3: AI Outfit Positioning (Main Feature!)

  Page: http://localhost:8556/outfit-builder.html

  1. Verify mannequin auto-loads:
    - Page loads with mannequin already visible
    - Placeholder text: "Click items from the right to add →"
  2. Add clothing items (try to get a complete outfit):
    - Add a top (t-shirt/shirt)
    - Add bottoms (pants/trousers)
    - Add outerwear (jacket/cardigan)
    - Add shoes
    - Expected: Placeholder disappears when first item added
    - Expected: Items appear scattered randomly on canvas
  3. Click "✨ Dress with AI":
    - Expected: Modal shows "AI is positioning clothing on
  body..."
    - Expected: After 60-90 seconds:
        - Tops positioned on upper body
      - Bottoms positioned on lower body
      - Shoes positioned at feet
      - Outerwear layered on top (higher z-index)
      - Proper layering (jacket OVER shirt)
    - Expected: Toast shows explanation
  4. Verify proper positioning:
    - Check that jacket/cardigan is visually on top of shirt
    - Check that pants are on lower body area
    - Check that shoes are near bottom

  Success Criteria:
  - ✅ Mannequin loads automatically
  - ✅ Items added to canvas
  - ✅ AI positions items correctly by category
  - ✅ Proper Z-ordering (outerwear > tops > bottoms > shoes)

  ---
  TEST 4: Outfit Rating

  Page: http://localhost:8556/outfit-builder.html

  1. With items on canvas from Test 3:
    - Click "⭐ Rate Outfit" button
    - Expected: Modal shows "AI is rating your outfit..."
    - Expected: After 10-20 seconds, see:
        - Star rating (⭐⭐⭐⭐⭐)
      - Numeric score (X/10)
      - Feedback paragraph
      - ✅ Strengths list
      - 💡 Improvements list

  Success Criteria:
  - ✅ Rating displayed (0-10)
  - ✅ Feedback is relevant to selected items
  - ✅ Strengths and improvements make sense

  ---
  TEST 5: AI Outfit Suggestions

  Page: http://localhost:8556/ai-stylist.html

  1. Navigate to AI Stylist page:
    - Should have at least 5+ items in wardrobe
  2. Test outfit suggestion:
    - Select occasion: "Casual"
    - Select weather: "Mild (15-20°C)"
    - Click "✨ Suggest Outfit"
    - Expected: Button shows "⏳ Thinking..."
    - Expected: After 10-20 seconds, see:
        - 3-5 selected items displayed with images
      - Reasoning why they work together
      - Styling tips list
  3. Create outfit from suggestion:
    - Click "📸 Create Outfit on Canvas"
    - Expected: Redirects to outfit builder
    - Expected: Selected items auto-added to canvas
    - Expected: Toast: "AI-suggested items added! Click Dress
   with AI to position them"
  4. Position suggested outfit:
    - Click "✨ Dress with AI"
    - Verify items positioned correctly

  Success Criteria:
  - ✅ AI suggests appropriate items for occasion/weather
  - ✅ Reasoning makes sense
  - ✅ Items transfer to outfit builder
  - ✅ Items can be positioned with AI

  ---
  TEST 6: Color Coordination Analysis

  Page: http://localhost:8556/ai-stylist.html

  1. Scroll to "Color Coordination" section:
    - See all wardrobe items as small cards
  2. Select items to analyze:
    - Click 2-3 items (they should highlight with border)
    - Expected: "🎨 Analyze Colors" button becomes enabled
  3. Analyze colors:
    - Click "🎨 Analyze Colors"
    - Expected: Button shows "⏳ Analyzing..."
    - Expected: After 5-15 seconds, see:
        - Compatibility score (0-100) with color coding
      - Green (80+), Yellow (60-79), Red (<60)
      - Analysis paragraph
      - Suggestions list

  Success Criteria:
  - ✅ Score displayed correctly
  - ✅ Analysis mentions specific colors
  - ✅ Suggestions are actionable

  ---
  TEST 7: Style Matching

  Page: http://localhost:8556/item-detail.html?id=1

  1. Navigate to any item:
    - Click "✨ Find Matching Items" button
    - Expected: Button shows "⏳ Finding matches..."
    - Expected: After 10-20 seconds:
        - Section appears with 5-6 matching items
      - Each shows match score (0-100) color-coded
      - Reason why it matches
      - Can click item to view details

  Success Criteria:
  - ✅ Matching items displayed
  - ✅ Scores make sense (similar colors/styles score higher)
  - ✅ Reasons are logical

  ---
  TEST 8: Wardrobe Gap Analysis

  Page: http://localhost:8556/ai-stylist.html

  1. Scroll to "Wardrobe Analysis":
    - Click "🔍 Analyze My Wardrobe"
    - Expected: Button shows "⏳ Analyzing..."
    - Expected: After 15-30 seconds, see:
        - Overall Assessment card
      - Suggested Items to Add card
      - What You Have Covered card
      - Wardrobe Tips card

  Success Criteria:
  - ✅ Assessment mentions wardrobe size and variety
  - ✅ Missing items are practical suggestions
  - ✅ Strengths accurately reflect category distribution
  - ✅ Tips are personalized

  ---
  TEST 9: Default Mannequin & Reset

  Page: http://localhost:8556/outfit-builder.html

  1. Test mannequin features:
    - Mannequin should be visible on page load
    - Add some items
    - Click "📸 Upload Your Photo" and upload a real body
  photo
    - Expected: Photo replaces mannequin
    - Click "🧍 Reset Mannequin"
    - Expected: Mannequin returns
    - Expected: Toast: "Mannequin reset!"
  2. Test clear all:
    - With items on canvas, click "Clear All"
    - Expected: All items removed
    - Expected: Mannequin resets
    - Expected: Placeholder reappears

  Success Criteria:
  - ✅ Can switch between custom photo and mannequin
  - ✅ Clear all resets everything properly

  ---
  TEST 10: Voice Input Support

  Page: Any page with 🎤 button (wardrobe, upload, chat, etc.)

  1. Test voice input availability:
    - Look for floating 🎤 button (bottom right)
    - Or 🎤 button in chat interface
    - Expected: Button visible on all main pages
  2. Test voice commands:
    - Click 🎤 microphone button
    - Expected: Button turns red (🔴), toast shows "Listening..."
    - Say: "Show my tops"
    - Expected: Navigates to wardrobe filtered by Tops
  3. Test chat voice input:
    - Go to outfit-builder.html or ai-stylist.html
    - Open chat widget (💬)
    - Click 🎤 in chat input area
    - Say: "Create a casual outfit"
    - Expected: Text appears in chat input, AI responds
  4. Test various voice commands:
    - "Add a new item" → Goes to upload.html
    - "Show my blue items" → Wardrobe with search
    - "Create an outfit" → Goes to outfit-builder.html
    - "Show analytics" → Goes to analytics.html
    - "Find my shoes" → Wardrobe filtered by Shoes
  5. Test error handling:
    - Click 🎤, don't speak (silence)
    - Expected: Timeout, button returns to normal
    - Deny microphone permission
    - Expected: Toast shows permission error
  6. Test stop recording:
    - Click 🎤 to start
    - Click 🔴 (red button) to stop
    - Expected: Stops listening, processes partial input

  Success Criteria:
  - ✅ Voice button visible on all pages
  - ✅ Recording starts/stops correctly
  - ✅ Transcription appears in chat or executes command
  - ✅ Voice commands navigate correctly
  - ✅ Error messages clear and helpful
  - ✅ Works in Chrome/Edge (primary browsers)

  Browser Compatibility:
  - ✅ Chrome/Edge: Full support
  - ⚠️ Firefox: Limited, may require flags
  - ⚠️ Safari: Requires explicit permissions

  ---
  TEST 11: End-to-End Workflow

  Complete user journey:

  1. Upload clothing → http://localhost:8556/upload.html
    - Upload 5-10 items with different categories
    - Verify background removal + AI analysis
  2. Get outfit suggestion →
  http://localhost:8556/ai-stylist.html
    - Select occasion + weather
    - Get AI suggestion
    - Click "Create Outfit on Canvas"
  3. Position with AI →
  http://localhost:8556/outfit-builder.html
    - Auto-loaded items appear
    - Click "✨ Dress with AI"
    - Verify proper positioning and layering
  4. Rate the outfit:
    - Click "⭐ Rate Outfit"
    - Review feedback
  5. Save outfit:
    - Click "💾 Save Outfit"
    - Name it and save
  6. View in outfits → http://localhost:8556/outfits.html
    - Verify saved outfit appears

  Success Criteria:
  - ✅ Complete flow works without errors
  - ✅ All AI features integrate smoothly
  - ✅ Data persists across pages

  ---
  PERFORMANCE BENCHMARKS

  Expected timings (CPU-only):
  - Background removal: 5-10 seconds
  - Image analysis (llava:7b): 60-90 seconds
  - Enhanced brand (llava:13b): 2-3 minutes
  - Outfit positioning (Dress AI): 60-90 seconds
  - Outfit suggestions: 10-20 seconds
  - Color analysis: 5-15 seconds
  - Style matching: 10-20 seconds
  - Wardrobe gaps: 15-30 seconds
  - Outfit rating: 10-20 seconds

  ---
  COMMON ISSUES & FIXES

  AI not available:
  docker ps | grep ollama
  ollama list
  # If missing: ollama pull llava:7b llama3.2:3b

  Background removal fails:
  docker ps | grep rembg
  docker compose up -d rembg
  curl http://localhost:5000  # Should return 200

  Server not responding:
  # Restart server
  cd mockups/api && go run main.go

  Slow performance:
  - Expected on CPU-only
  - llava:7b is faster than llava:13b
  - Consider using llava:7b for all tasks

  ---
  TEST COMPLETION CHECKLIST

  Core Features:
  - [x] Multi-image upload works
  - [x] Background removal works
  - [ ] AI detects category/color/brand
  - [ ] Enhanced brand detection (13b) works
  - [ ] Mannequin auto-loads
  - [ ] Items can be added to canvas
  - [ ] "Dress with AI" positions items correctly
  - [ ] Proper layering (outerwear over tops)
  - [ ] Outfit rating provides feedback
  - [ ] Outfit suggestions work
  - [ ] Color analysis scores items
  - [ ] Style matching finds similar items
  - [ ] Wardrobe gaps analysis complete
  - [ ] Can upload custom body photo
  - [ ] Can reset to mannequin
  - [ ] Clear all resets canvas
  - [ ] Data persists across pages
  - [ ] All pages navigate correctly

  Outfit Management:
  - [ ] Can edit existing outfits
  - [ ] Can duplicate outfits
  - [ ] Can delete outfits
  - [ ] Outfit search filters work
  - [ ] Outfit rating filters work
  - [ ] Outfit sorting works

  Analytics:
  - [ ] Analytics dashboard shows stats
  - [ ] Charts render correctly
  - [ ] Cost-per-wear calculation works
  - [ ] AI insights generate correctly
  - [ ] Most/least worn items displayed

  Calendar:
  - [ ] Can view calendar
  - [ ] Can plan outfits for dates
  - [ ] AI suggests outfits for events
  - [ ] Calendar events save properly

  Settings:
  - [ ] Can update preferences
  - [ ] Settings persist across pages
  - [ ] Can configure AI models
  - [ ] Can export/import data

  Sharing:
  - [ ] Can generate shareable links
  - [ ] Public outfit page works
  - [ ] Share modal displays correctly
  - [ ] Can copy share URL

  Chat Interface:
  - [ ] Chat UI works
  - [ ] Can request outfit via chat
  - [ ] AI responds with suggestions
  - [ ] Can create outfit from chat

  AI Enhancements:
  - [ ] AI can regenerate outfit variations
  - [ ] "Make more formal/casual" works
  - [ ] AI calendar suggestions work
  - [ ] AI analytics insights useful

  Voice Input:
  - [ ] Voice button visible on pages
  - [ ] Can start/stop recording
  - [ ] Speech transcription works
  - [ ] Voice commands navigate correctly
  - [ ] Chat accepts voice input
  - [ ] Error handling works (permissions, timeout)
  - [ ] Works in Chrome/Edge

  All tests passing = Ready for backend implementation! 🎉
