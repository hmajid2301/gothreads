/**
 * go-threads Mockup JavaScript
 * Throwaway prototype code - will be replaced with HTMX + Alpine + Svelte
 */

// ============================================================================
// AI API CONFIGURATION
// ============================================================================

const AI_API_URL = 'http://localhost:8556/api';
let aiAvailable = false;

async function checkAIStatus() {
  try {
    const resp = await fetch(`${AI_API_URL}/status`);
    if (resp.ok) {
      const status = await resp.json();
      aiAvailable = status.ready;
      console.log('AI Status:', status);
      updateAIStatusIndicator(status);
      return status;
    }
  } catch (e) {
    console.log('AI API not available:', e.message);
    aiAvailable = false;
  }
  updateAIStatusIndicator(null);
  return null;
}

function updateAIStatusIndicator(status) {
  const indicator = document.getElementById('ai-status');
  if (!indicator) return;

  if (status && status.ready) {
    indicator.innerHTML = `<span style="color: #27AE60;">● AI Ready</span>`;
    indicator.title = `Vision: ${status.vision_model}\nText: ${status.text_model}`;
  } else if (status) {
    indicator.innerHTML = `<span style="color: #F39C12;">● AI: Models loading</span>`;
    indicator.title = `Available: ${status.available_models?.join(', ') || 'none'}`;
  } else {
    indicator.innerHTML = `<span style="color: #95a5a6;">○ AI Offline</span>`;
    indicator.title = 'Start the API: cd mockups/api && go run main.go';
  }
}

async function analyzeImageWithAI(imageDataUrl, model = null) {
  if (!aiAvailable) {
    return null;
  }

  try {
    const requestBody = { image: imageDataUrl };
    if (model) {
      requestBody.model = model;
    }

    const resp = await fetch(`${AI_API_URL}/analyze`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(requestBody)
    });

    if (resp.ok) {
      return await resp.json();
    } else {
      const error = await resp.json();
      console.error('AI analysis error:', error);
      showToast(`AI error: ${error.error || 'Unknown error'}`, 'error');
    }
  } catch (e) {
    console.error('AI analysis failed:', e);
    showToast('AI analysis failed: ' + e.message, 'error');
  }
  return null;
}

async function generateTagsWithAI(description) {
  if (!aiAvailable) {
    return [];
  }

  try {
    const resp = await fetch(`${AI_API_URL}/tags`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ description })
    });

    if (resp.ok) {
      const data = await resp.json();
      return data.tags || [];
    }
  } catch (e) {
    console.error('Tag generation failed:', e);
  }
  return [];
}

async function dressWithAI(bodyImage, clothingItems, customPrompt = '') {
  if (!aiAvailable) {
    return null;
  }

  try {
    const resp = await fetch(`${AI_API_URL}/dress`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        body_image: bodyImage,
        clothing_items: clothingItems,
        prompt: customPrompt
      })
    });

    if (resp.ok) {
      return await resp.json();
    } else {
      const error = await resp.json();
      console.error('AI dress error:', error);
      showToast(`AI error: ${error.error || 'Unknown error'}`, 'error');
    }
  } catch (e) {
    console.error('AI dress failed:', e);
    showToast('AI dress failed: ' + e.message, 'error');
  }
  return null;
}

async function removeBackgroundWithAI(imageDataUrl) {
  try {
    const resp = await fetch(`${AI_API_URL}/remove-bg`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ image: imageDataUrl })
    });

    if (resp.ok) {
      const data = await resp.json();
      return data.image;
    } else {
      const error = await resp.json();
      console.error('Background removal error:', error);
      showToast(`Background removal failed: ${error.error || 'Unknown error'}`, 'error');
    }
  } catch (e) {
    console.error('Background removal failed:', e);
    showToast('Background removal service not available', 'error');
  }
  return null;
}

async function suggestOutfitWithAI(occasion, weather, items) {
  if (!aiAvailable) return null;

  try {
    const resp = await fetch(`${AI_API_URL}/suggest-outfit`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ occasion, weather, items })
    });

    if (resp.ok) {
      return await resp.json();
    } else {
      const error = await resp.json();
      console.error('Outfit suggestion error:', error);
      showToast(`AI error: ${error.error || 'Unknown error'}`, 'error');
    }
  } catch (e) {
    console.error('Outfit suggestion failed:', e);
    showToast('Outfit suggestion failed: ' + e.message, 'error');
  }
  return null;
}

async function analyzeColorMatch(items) {
  if (!aiAvailable) return null;

  try {
    const resp = await fetch(`${AI_API_URL}/color-match`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ items })
    });

    if (resp.ok) {
      return await resp.json();
    } else {
      const error = await resp.json();
      console.error('Color match error:', error);
    }
  } catch (e) {
    console.error('Color match failed:', e);
  }
  return null;
}

async function findStyleMatches(baseItem, items, maxItems = 5) {
  if (!aiAvailable) return null;

  try {
    const resp = await fetch(`${AI_API_URL}/style-match`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ base_item: baseItem, items, max_items: maxItems })
    });

    if (resp.ok) {
      return await resp.json();
    } else {
      const error = await resp.json();
      console.error('Style match error:', error);
    }
  } catch (e) {
    console.error('Style match failed:', e);
  }
  return null;
}

async function rateOutfitWithAI(items) {
  if (!aiAvailable) return null;

  try {
    const resp = await fetch(`${AI_API_URL}/rate-outfit`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ items })
    });

    if (resp.ok) {
      return await resp.json();
    } else {
      const error = await resp.json();
      console.error('Outfit rating error:', error);
    }
  } catch (e) {
    console.error('Outfit rating failed:', e);
  }
  return null;
}

async function analyzeWardrobeGaps(items) {
  if (!aiAvailable) return null;

  try {
    const resp = await fetch(`${AI_API_URL}/wardrobe-gaps`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ items })
    });

    if (resp.ok) {
      return await resp.json();
    } else {
      const error = await resp.json();
      console.error('Wardrobe gaps error:', error);
    }
  } catch (e) {
    console.error('Wardrobe gaps analysis failed:', e);
  }
  return null;
}

// ============================================================================
// DATA STORE (Mock localStorage persistence)
// ============================================================================

const STORAGE_KEY = 'gothreads_mockup_data';

const defaultData = {
  items: [
    // Bottoms
    { id: 1, name: 'Easies Navy Trousers', category: 'Bottoms', price: 119.00, color: 'Navy', brand: 'MR MARVIS', wearCount: 24, image: 'images/examples/easies_navy.avif' },
    { id: 2, name: 'Easies Green Trousers', category: 'Bottoms', price: 119.00, color: 'Green', brand: 'MR MARVIS', wearCount: 18, image: 'images/examples/easies_green.avif' },
    { id: 3, name: 'Easies Royal Blue', category: 'Bottoms', price: 119.00, color: 'Royal Blue', brand: 'MR MARVIS', wearCount: 12, image: 'images/examples/easies_royal_blue.avif' },
    { id: 4, name: 'Active Beige Trousers', category: 'Bottoms', price: 129.00, color: 'Beige', brand: 'MR MARVIS', wearCount: 8, image: 'images/examples/active_biege.avif' },
    { id: 5, name: 'Smart Navy Easy', category: 'Bottoms', price: 139.00, color: 'Navy', brand: 'MR MARVIS', wearCount: 6, image: 'images/examples/smart_navy_easy.avif' },
    // Tops
    { id: 6, name: 'Midweight Navy Tee', category: 'Tops', price: 59.00, color: 'Navy', brand: 'MR MARVIS', wearCount: 15, image: 'images/examples/midweight_navy.avif' },
    { id: 7, name: 'Midweight White Tee', category: 'Tops', price: 59.00, color: 'White', brand: 'MR MARVIS', wearCount: 22, image: 'images/examples/midweight_white.avif' },
    { id: 8, name: 'Easy Shirt Deep Blue', category: 'Tops', price: 89.00, color: 'Deep Blue', brand: 'MR MARVIS', wearCount: 10, image: 'images/examples/easy_shirt_deep.avif' },
    { id: 9, name: 'Easy Shirt Green', category: 'Tops', price: 89.00, color: 'Green', brand: 'MR MARVIS', wearCount: 7, image: 'images/examples/easy_shirt_green.avif' },
    { id: 10, name: 'Classic Polo', category: 'Tops', price: 79.00, color: 'Navy', brand: 'MR MARVIS', wearCount: 11, image: 'images/examples/76fa61f499afa36c2f955ebac85d1b470099a50c-2000x2999.avif' },
    { id: 11, name: 'Oxford Shirt', category: 'Tops', price: 99.00, color: 'Light Blue', brand: 'MR MARVIS', wearCount: 9, image: 'images/examples/93c26d039e3cb4889ee685e0493730a9faa3e1bf-2000x3000.avif' },
    { id: 12, name: 'Linen Shirt', category: 'Tops', price: 109.00, color: 'White', brand: 'MR MARVIS', wearCount: 5, image: 'images/examples/a5dec65faf4b7dca5b50eb711f669c709db40b55-2000x3000.avif' },
    { id: 13, name: 'Casual Shirt', category: 'Tops', price: 89.00, color: 'Olive', brand: 'MR MARVIS', wearCount: 7, image: 'images/examples/caa5c91e1f68eb29cdcfe46aad34d867bc1d8409-2000x3000.avif' },
    { id: 14, name: 'Summer Tee', category: 'Tops', price: 49.00, color: 'Sage', brand: 'MR MARVIS', wearCount: 14, image: 'images/examples/2449ab77bfe241627cd7caff0c745f9382f98db3-1050x1575.avif' },
    // Outerwear
    { id: 15, name: 'Shacket', category: 'Outerwear', price: 169.00, color: 'Brown', brand: 'MR MARVIS', wearCount: 4, image: 'images/examples/shacket.avif' },
    { id: 16, name: 'Shawl Cardigan Navy', category: 'Outerwear', price: 149.00, color: 'Navy', brand: 'MR MARVIS', wearCount: 6, image: 'images/examples/shawl_navy_cardigan.jpg' },
    { id: 17, name: 'Shawl Cardigan Beige', category: 'Outerwear', price: 149.00, color: 'Beige', brand: 'MR MARVIS', wearCount: 3, image: 'images/examples/shawl_biee_cardigan.jpg' },
  ],
  outfits: [
    { id: 1, name: 'Smart Casual', itemIds: [1, 7], wearCount: 5, rating: 4.5, positions: [{id: 1, x: 120, y: 180, z: 0}, {id: 7, x: 100, y: 30, z: 1}] },
    { id: 2, name: 'Summer Vibes', itemIds: [3, 6], wearCount: 8, rating: 5.0, positions: [{id: 3, x: 100, y: 180, z: 0}, {id: 6, x: 110, y: 30, z: 1}] },
    { id: 3, name: 'Layered Look', itemIds: [5, 8, 15], wearCount: 3, rating: 4.0, positions: [{id: 5, x: 100, y: 200, z: 0}, {id: 8, x: 90, y: 80, z: 1}, {id: 15, x: 80, y: 20, z: 2}] },
    { id: 4, name: 'Cozy Office', itemIds: [1, 10, 16], wearCount: 2, rating: 4.8, positions: [{id: 1, x: 100, y: 200, z: 0}, {id: 10, x: 90, y: 80, z: 1}, {id: 16, x: 80, y: 20, z: 2}] },
    { id: 5, name: 'Weekend Relaxed', itemIds: [4, 14, 17], wearCount: 4, rating: 4.2, positions: [{id: 4, x: 100, y: 200, z: 0}, {id: 14, x: 90, y: 60, z: 1}, {id: 17, x: 85, y: 10, z: 2}] },
  ],
  wearHistory: [],
};

function getData() {
  const stored = localStorage.getItem(STORAGE_KEY);
  if (stored) {
    return JSON.parse(stored);
  }
  return defaultData;
}

function saveData(data) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
}

function clearAllData() {
  localStorage.removeItem(STORAGE_KEY);
  closeModal();
  showToast('All data cleared successfully!', 'success');
  setTimeout(() => {
    window.location.reload();
  }, 1000);
}

// Global state
let appData = getData();
let currentOutfitItems = [];
let selectedCanvasItem = null;

// ============================================================================
// TOAST NOTIFICATIONS
// ============================================================================

function showToast(message, type = 'info') {
  let toast = document.getElementById('toast');
  if (!toast) {
    toast = document.createElement('div');
    toast.id = 'toast';
    toast.className = 'toast';
    document.body.appendChild(toast);
  }

  toast.textContent = message;
  toast.className = `toast show ${type}`;

  setTimeout(() => {
    toast.classList.remove('show');
  }, 3000);
}

// ============================================================================
// MODAL SYSTEM
// ============================================================================

function showModal(title, content, actions = []) {
  // Remove existing modal
  const existing = document.getElementById('modal-overlay');
  if (existing) existing.remove();

  const overlay = document.createElement('div');
  overlay.id = 'modal-overlay';
  overlay.className = 'modal-overlay';
  overlay.innerHTML = `
    <div class="modal">
      <div class="modal-header">
        <h2>${title}</h2>
        <button class="modal-close" onclick="closeModal()">&times;</button>
      </div>
      <div class="modal-body">${content}</div>
      <div class="modal-actions">
        ${actions.map(a => `<button class="btn ${a.primary ? 'btn-primary' : 'btn-secondary'}" onclick="${a.onclick}">${a.label}</button>`).join('')}
      </div>
    </div>
  `;

  document.body.appendChild(overlay);
  overlay.addEventListener('click', (e) => {
    if (e.target === overlay) closeModal();
  });
}

function closeModal() {
  const modal = document.getElementById('modal-overlay');
  if (modal) modal.remove();
}

// ============================================================================
// UPLOAD HANDLING
// ============================================================================

function initUpload() {
  const uploadArea = document.querySelector('.upload-area');
  const uploadInput = document.getElementById('upload-input');

  if (!uploadArea || !uploadInput) return;

  // Clear database button
  const clearDbBtn = document.getElementById('clear-db-btn');
  if (clearDbBtn) {
    clearDbBtn.onclick = () => {
      showModal('Clear All Data?', `
        <p>This will permanently delete all clothing items and outfits from your wardrobe.</p>
        <p style="color: #E74C3C; font-weight: bold;">This action cannot be undone!</p>
      `, [
        { label: 'Cancel', onclick: 'closeModal()' },
        { label: 'Clear All Data', primary: true, onclick: 'clearAllData()' }
      ]);
    };
  }

  // Drag and drop
  uploadArea.addEventListener('dragover', (e) => {
    e.preventDefault();
    uploadArea.classList.add('drag-over');
  });

  uploadArea.addEventListener('dragleave', () => {
    uploadArea.classList.remove('drag-over');
  });

  uploadArea.addEventListener('drop', (e) => {
    e.preventDefault();
    uploadArea.classList.remove('drag-over');
    const files = e.dataTransfer.files;
    if (files.length > 0) {
      handleMultipleFileUploads(files);
    }
  });

  // File input
  uploadInput.addEventListener('change', (e) => {
    if (e.target.files.length > 0) {
      handleMultipleFileUploads(e.target.files);
    }
  });
}

async function handleMultipleFileUploads(files) {
  const fileArray = Array.from(files);

  if (fileArray.length === 0) return;

  // Show modal with progress
  showBatchUploadModal(fileArray.length);

  let successCount = 0;
  let failCount = 0;

  for (let i = 0; i < fileArray.length; i++) {
    const file = fileArray[i];
    updateBatchUploadProgress(i + 1, fileArray.length, file.name);

    const result = await handleSingleFileUpload(file, false); // Don't redirect
    if (result) {
      successCount++;
    } else {
      failCount++;
    }
  }

  closeModal();

  if (successCount > 0) {
    showToast(`Successfully uploaded ${successCount} item${successCount > 1 ? 's' : ''}! Review them below.`, 'success');
    setTimeout(() => {
      window.location.href = 'wardrobe.html?recent=true';
    }, 1500);
  }

  if (failCount > 0) {
    showToast(`Failed to upload ${failCount} item${failCount > 1 ? 's' : ''}`, 'error');
  }
}

function showBatchUploadModal(totalCount) {
  showModal('Uploading Items', `
    <div style="text-align: center; padding: 2rem;">
      <div class="spinner"></div>
      <p id="batch-upload-status" style="margin-top: 1rem; color: var(--text-neutral);">
        Processing 0 of ${totalCount} items...
      </p>
      <p id="batch-upload-file" style="margin-top: 0.5rem; font-size: 0.875rem; color: var(--text-neutral);"></p>
      <div class="progress-bar" style="margin-top: 1rem;">
        <div class="progress-fill" id="batch-upload-progress" style="width: 0%"></div>
      </div>
    </div>
  `, []);
}

function updateBatchUploadProgress(current, total, filename) {
  const statusEl = document.getElementById('batch-upload-status');
  const fileEl = document.getElementById('batch-upload-file');
  const progressEl = document.getElementById('batch-upload-progress');

  if (statusEl) statusEl.textContent = `Processing ${current} of ${total} items...`;
  if (fileEl) fileEl.textContent = filename;
  if (progressEl) progressEl.style.width = `${(current / total) * 100}%`;
}

async function handleSingleFileUpload(file, shouldRedirect = true) {
  // Validate file type
  const validTypes = ['image/jpeg', 'image/png', 'image/webp', 'image/avif'];
  if (!validTypes.includes(file.type)) {
    showToast('Please upload a JPEG, PNG, WebP, or AVIF image', 'error');
    return false;
  }

  // Validate file size (10MB)
  if (file.size > 10 * 1024 * 1024) {
    showToast('File too large. Maximum size is 10MB', 'error');
    return false;
  }

  return new Promise((resolve) => {
    const reader = new FileReader();
    reader.onload = async function(e) {
      let imageData = e.target.result;

      try {
        // Remove background first (always)
        updateProcessingStatus('Removing background...');
        const bgRemovedImage = await removeBackgroundWithAI(imageData);
        if (bgRemovedImage) {
          imageData = bgRemovedImage;
          console.log('Background removed successfully');
        } else {
          console.log('Background removal failed, using original image');
        }

        // Try AI analysis if available
        let aiResult = null;
        if (aiAvailable) {
          updateProcessingStatus('Analyzing clothing with AI...');
          aiResult = await analyzeImageWithAI(imageData);
          if (aiResult) {
            console.log('AI Analysis Result:', aiResult);
          }
        }

        const newItem = {
          id: Date.now() + Math.random(), // Ensure unique IDs for batch uploads
          name: aiResult?.description?.substring(0, 50) || file.name.replace(/\.[^/.]+$/, '').replace(/[-_]/g, ' '),
          category: aiResult?.category || 'Uncategorized',
          price: 0,
          color: aiResult?.color || '',
          brand: aiResult?.brand || '',
          wearCount: 0,
          image: imageData,
          tags: aiResult?.tags || [],
          aiAnalysis: aiResult?.raw_response || null,
          uploadedAt: new Date().toISOString(),
        };

        appData.items.push(newItem);
        saveData(appData);

        if (shouldRedirect) {
          closeModal();

          if (aiResult) {
            const brandText = aiResult.brand ? ` by ${aiResult.brand}` : '';
            showToast(`AI detected: ${aiResult.color} ${aiResult.category}${brandText}`, 'success');
          } else {
            showToast('Item uploaded! Edit details below.', 'success');
          }

          // Redirect to edit page
          setTimeout(() => {
            window.location.href = `item-detail.html?id=${newItem.id}`;
          }, 1000);
        }

        resolve(true);
      } catch (error) {
        console.error('Upload error:', error);
        resolve(false);
      }
    };

    reader.onerror = () => resolve(false);
    reader.readAsDataURL(file);
  });
}

function showProcessingModal(statusText = 'Uploading and analyzing image...') {
  showModal('Processing Image', `
    <div style="text-align: center; padding: 2rem;">
      <div class="spinner"></div>
      <p id="processing-status" style="margin-top: 1rem; color: var(--text-neutral);">${statusText}</p>
      <div class="progress-bar" style="margin-top: 1rem;">
        <div class="progress-fill" id="upload-progress"></div>
      </div>
    </div>
  `, []);

  // Animate progress (slower for AI)
  let progress = 0;
  const progressBar = document.getElementById('upload-progress');
  const interval = setInterval(() => {
    progress += 2;
    if (progressBar) progressBar.style.width = `${Math.min(progress, 90)}%`;
    if (progress >= 90) clearInterval(interval);
  }, 100);
}

function updateProcessingStatus(text) {
  const status = document.getElementById('processing-status');
  if (status) status.textContent = text;
}

// ============================================================================
// WARDROBE GRID
// ============================================================================

function renderWardrobe(containerId, filterCategory = 'all') {
  const container = document.getElementById(containerId);
  if (!container) return;

  // Check if we should show recent uploads
  const urlParams = new URLSearchParams(window.location.search);
  const showRecent = urlParams.get('recent') === 'true';

  let items = appData.items;
  if (filterCategory !== 'all') {
    items = items.filter(item => item.category === filterCategory);
  }

  // Sort by upload date if showing recent
  if (showRecent) {
    items = items.sort((a, b) => new Date(b.uploadedAt || 0) - new Date(a.uploadedAt || 0));
  }

  container.innerHTML = '';

  if (items.length === 0) {
    container.innerHTML = `
      <div style="grid-column: 1/-1; text-align: center; padding: 3rem; color: var(--text-neutral);">
        <div style="font-size: 3rem; margin-bottom: 1rem;">👕</div>
        <p>No items found. <a href="upload.html" style="color: var(--primary);">Upload your first item!</a></p>
      </div>
    `;
    return;
  }

  // Show banner if recent uploads
  if (showRecent) {
    const banner = document.createElement('div');
    banner.style.cssText = 'grid-column: 1/-1; background: var(--warm-bg); padding: 1rem; border-radius: 0.5rem; border: 1px solid var(--primary); margin-bottom: 1rem;';
    banner.innerHTML = `
      <p style="margin: 0; color: var(--text-neutral);">
        ✨ Recently uploaded items shown first. Click any item to review and edit AI-detected details.
      </p>
    `;
    container.appendChild(banner);
  }

  items.forEach(item => {
    const card = document.createElement('div');
    card.className = 'item-card';
    card.onclick = () => window.location.href = `item-detail.html?id=${item.id}`;

    const costPerWear = item.wearCount > 0 ? (item.price / item.wearCount).toFixed(2) : item.price.toFixed(2);

    // Highlight AI-detected info
    const aiInfo = item.aiAnalysis ? `
      <div style="font-size: 0.75rem; color: var(--primary); margin-top: 0.25rem;">
        ✨ ${item.color}${item.brand ? ` • ${item.brand}` : ''}
      </div>
    ` : '';

    card.innerHTML = `
      <img src="${item.image}" alt="${item.name}" class="item-image" loading="lazy">
      <div class="item-info">
        <div class="item-name">${item.name}</div>
        <div class="item-meta">${item.category} • $${item.price.toFixed(2)}</div>
        ${aiInfo}
        <div class="item-stats">
          <span title="Times worn">👔 ${item.wearCount}</span>
          <span title="Cost per wear">💰 $${costPerWear}</span>
        </div>
      </div>
    `;

    container.appendChild(card);
  });
}

function initWardrobeFilter() {
  const filter = document.getElementById('category-filter');
  if (!filter) return;

  filter.addEventListener('change', (e) => {
    renderWardrobe('wardrobe-grid', e.target.value);
  });
}

// ============================================================================
// OUTFIT BUILDER (CANVAS)
// ============================================================================

function initOutfitBuilder() {
  const canvas = document.querySelector('.outfit-canvas');
  const picker = document.getElementById('item-picker');

  if (!canvas || !picker) return;

  // Auto-load default mannequin on page load
  window.mannequinAutoLoaded = true;
  useDefaultMannequin();

  // Render item picker with category tabs
  renderItemPicker(picker);

  // Setup canvas interactions
  canvas.addEventListener('click', (e) => {
    if (e.target === canvas) {
      deselectAllItems();
    }
  });

  // Keyboard shortcuts
  document.addEventListener('keydown', (e) => {
    if (!selectedCanvasItem) return;

    if (e.key === 'Delete' || e.key === 'Backspace') {
      removeFromCanvas(selectedCanvasItem.dataset.itemId);
    } else if (e.key === 'ArrowUp' && e.shiftKey) {
      bringForward(selectedCanvasItem);
    } else if (e.key === 'ArrowDown' && e.shiftKey) {
      sendBackward(selectedCanvasItem);
    }
  });

  // Setup body upload
  const bodyUploadInput = document.getElementById('body-upload-input');
  if (bodyUploadInput) {
    bodyUploadInput.addEventListener('change', (e) => {
      if (e.target.files.length > 0) {
        handleBodyUpload(e.target.files[0]);
      }
    });
  }

  // Setup save outfit button
  const saveBtn = document.getElementById('save-outfit');
  if (saveBtn) {
    saveBtn.onclick = saveOutfit;
  }

  // Setup AI dress button
  const dressBtn = document.getElementById('dress-with-ai');
  if (dressBtn) {
    dressBtn.onclick = handleDressWithAI;
  }

  // Rate Outfit button
  const rateBtn = document.getElementById('rate-outfit-btn');
  if (rateBtn) {
    rateBtn.onclick = handleRateOutfit;
  }

  // Check if outfit was suggested from AI Stylist
  const urlParams = new URLSearchParams(window.location.search);
  if (urlParams.get('suggested') === 'true') {
    const suggestedItems = JSON.parse(sessionStorage.getItem('suggested_outfit_items') || '[]');
    if (suggestedItems.length > 0) {
      // Auto-add suggested items to canvas
      setTimeout(() => {
        suggestedItems.forEach(itemId => {
          const item = appData.items.find(i => i.id === itemId);
          if (item) {
            addToCanvas(item.id);
          }
        });
        sessionStorage.removeItem('suggested_outfit_items');
        showToast('AI-suggested items added! Click "Dress with AI" to position them.', 'success');
      }, 500);
    }
  }
}

function handleBodyUpload(file) {
  const validTypes = ['image/jpeg', 'image/png', 'image/webp', 'image/avif'];
  if (!validTypes.includes(file.type)) {
    showToast('Please upload a JPEG, PNG, WebP, or AVIF image', 'error');
    return;
  }

  const reader = new FileReader();
  reader.onload = function(e) {
    const canvas = document.querySelector('.outfit-canvas');
    if (canvas) {
      canvas.style.backgroundImage = `url(${e.target.result})`;
      showToast('Body image set as background', 'success');
    }
  };
  reader.readAsDataURL(file);
}

function useDefaultMannequin() {
  // Create a simple SVG mannequin as data URL
  const mannequinSVG = `
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 200 500" width="200" height="500">
  <!-- Head -->
  <ellipse cx="100" cy="40" rx="25" ry="30" fill="#E8DDD3" stroke="#C4B5A0" stroke-width="2"/>

  <!-- Neck -->
  <rect x="90" y="65" width="20" height="15" fill="#E8DDD3" stroke="#C4B5A0" stroke-width="1"/>

  <!-- Shoulders and Torso -->
  <path d="M 60 80 L 60 240 L 140 240 L 140 80 Z" fill="#F5F0E8" stroke="#C4B5A0" stroke-width="2"/>
  <line x1="60" y1="80" x2="30" y2="100" stroke="#C4B5A0" stroke-width="2"/>
  <line x1="140" y1="80" x2="170" y2="100" stroke="#C4B5A0" stroke-width="2"/>

  <!-- Arms -->
  <rect x="25" y="100" width="12" height="140" fill="#E8DDD3" stroke="#C4B5A0" stroke-width="2" rx="6"/>
  <rect x="163" y="100" width="12" height="140" fill="#E8DDD3" stroke="#C4B5A0" stroke-width="2" rx="6"/>

  <!-- Waist line -->
  <line x1="60" y1="180" x2="140" y2="180" stroke="#C4B5A0" stroke-width="1" stroke-dasharray="5,5"/>

  <!-- Hips/Legs -->
  <path d="M 70 240 L 65 380 L 85 380 L 85 240 Z" fill="#E8DDD3" stroke="#C4B5A0" stroke-width="2"/>
  <path d="M 130 240 L 135 380 L 115 380 L 115 240 Z" fill="#E8DDD3" stroke="#C4B5A0" stroke-width="2"/>

  <!-- Feet base -->
  <ellipse cx="75" cy="385" rx="15" ry="8" fill="#C4B5A0" opacity="0.3"/>
  <ellipse cx="125" cy="385" rx="15" ry="8" fill="#C4B5A0" opacity="0.3"/>

  <!-- Guidelines -->
  <text x="100" y="20" text-anchor="middle" font-size="12" fill="#999" font-family="Arial">Mannequin</text>
</svg>
  `.trim();

  const canvas = document.querySelector('.outfit-canvas');
  if (canvas) {
    const dataUrl = 'data:image/svg+xml;base64,' + btoa(mannequinSVG);
    canvas.style.backgroundImage = `url(${dataUrl})`;
    canvas.style.backgroundSize = 'contain';
    canvas.style.backgroundPosition = 'center';
    canvas.style.backgroundRepeat = 'no-repeat';

    // Only show toast if manually clicked (not auto-load)
    if (window.mannequinAutoLoaded !== true) {
      showToast('Mannequin reset! Add items and try "Dress with AI"!', 'success');
    }
  }
}

function removeBodyBackground() {
  const canvas = document.querySelector('.outfit-canvas');
  if (canvas) {
    canvas.style.backgroundImage = 'none';
    showToast('Background removed', 'info');
  }
}

function clearCanvas() {
  const canvas = document.querySelector('.outfit-canvas');
  if (canvas) {
    currentOutfitItems = [];
    canvas.innerHTML = `
      <div id="canvas-placeholder" style="position: absolute; top: 50%; left: 50%; transform: translate(-50%, -50%); text-align: center; color: var(--text-neutral); pointer-events: none;">
        <div style="font-size: 3rem; margin-bottom: 0.5rem;">👕</div>
        <p style="margin: 0;">Click items from the right to add →</p>
        <p style="margin: 0.5rem 0 0 0; font-size: 0.875rem;">Then click "✨ Dress with AI" to position them</p>
      </div>
    `;

    // Reset mannequin
    window.mannequinAutoLoaded = true;
    useDefaultMannequin();

    renderPickerItems(document.querySelector('.picker-tab.active')?.dataset.category || 'all');
    showToast('Canvas cleared and mannequin reset', 'info');
  }
}

function renderItemPicker(container) {
  const categories = ['All', 'Tops', 'Bottoms', 'Shoes', 'Outerwear', 'Accessories'];

  container.innerHTML = `
    <div class="picker-tabs">
      ${categories.map((cat, i) => `
        <button class="picker-tab ${i === 0 ? 'active' : ''}" data-category="${cat.toLowerCase()}">${cat}</button>
      `).join('')}
    </div>
    <div class="picker-items" id="picker-items"></div>
  `;

  // Tab switching
  container.querySelectorAll('.picker-tab').forEach(tab => {
    tab.addEventListener('click', () => {
      container.querySelectorAll('.picker-tab').forEach(t => t.classList.remove('active'));
      tab.classList.add('active');
      renderPickerItems(tab.dataset.category);
    });
  });

  renderPickerItems('all');
}

function renderPickerItems(category) {
  const container = document.getElementById('picker-items');
  if (!container) return;

  let items = appData.items;
  if (category !== 'all') {
    items = items.filter(item => item.category.toLowerCase() === category);
  }

  container.innerHTML = '';

  items.forEach(item => {
    const isInOutfit = currentOutfitItems.some(i => i.id === item.id);

    const pickerItem = document.createElement('div');
    pickerItem.className = `picker-item ${isInOutfit ? 'in-outfit' : ''}`;
    pickerItem.onclick = () => {
      if (isInOutfit) {
        removeFromCanvas(item.id);
      } else {
        addToCanvas(item);
      }
    };

    pickerItem.innerHTML = `
      <img src="${item.image}" alt="${item.name}">
      <div class="picker-item-info">
        <div class="picker-item-name">${item.name}</div>
        <div class="picker-item-category">${item.category}</div>
      </div>
      <div class="picker-item-status">${isInOutfit ? '✓' : '+'}</div>
    `;

    container.appendChild(pickerItem);
  });
}

function addToCanvas(item) {
  const canvas = document.querySelector('.outfit-canvas');
  if (!canvas) return;

  // Check if already in outfit
  if (currentOutfitItems.some(i => i.id === item.id)) {
    showToast('Item already in outfit', 'info');
    return;
  }

  const canvasRect = canvas.getBoundingClientRect();
  const x = Math.random() * (canvasRect.width - 150) + 50;
  const y = Math.random() * (canvasRect.height - 150) + 50;
  const z = currentOutfitItems.length;

  const canvasItem = document.createElement('div');
  canvasItem.className = 'canvas-item';
  canvasItem.dataset.itemId = item.id;
  canvasItem.style.left = `${x}px`;
  canvasItem.style.top = `${y}px`;
  canvasItem.style.zIndex = z;

  canvasItem.innerHTML = `
    <img src="${item.image}" alt="${item.name}" draggable="false">
    <div class="canvas-item-controls">
      <button onclick="bringForward(this.closest('.canvas-item'))" title="Bring forward">↑</button>
      <button onclick="sendBackward(this.closest('.canvas-item'))" title="Send backward">↓</button>
      <button onclick="removeFromCanvas(${item.id})" title="Remove">×</button>
    </div>
  `;

  // Setup dragging
  setupCanvasItemDrag(canvasItem);

  // Click to select and bring to front
  canvasItem.addEventListener('click', (e) => {
    if (e.target.tagName !== 'BUTTON') {
      selectCanvasItem(canvasItem);
      // Bring to front when clicked
      const allItems = Array.from(canvas.querySelectorAll('.canvas-item'));
      const maxZ = Math.max(...allItems.map(el => parseInt(el.style.zIndex) || 0));
      canvasItem.style.zIndex = maxZ + 1;
    }
  });

  canvas.appendChild(canvasItem);
  currentOutfitItems.push({ ...item, x, y, z });

  // Hide placeholder when first item is added
  const placeholder = document.getElementById('canvas-placeholder');
  if (placeholder) {
    placeholder.style.display = 'none';
  }

  // Update picker
  renderPickerItems(document.querySelector('.picker-tab.active')?.dataset.category || 'all');

  showToast(`Added ${item.name}`, 'success');
}

function removeFromCanvas(itemId) {
  itemId = parseInt(itemId);
  const canvasItem = document.querySelector(`.canvas-item[data-item-id="${itemId}"]`);
  if (canvasItem) {
    canvasItem.remove();
  }

  currentOutfitItems = currentOutfitItems.filter(i => i.id !== itemId);
  selectedCanvasItem = null;

  // Update picker
  renderPickerItems(document.querySelector('.picker-tab.active')?.dataset.category || 'all');

  showToast('Item removed', 'info');
}

function selectCanvasItem(element) {
  deselectAllItems();
  element.classList.add('selected');
  selectedCanvasItem = element;
}

function deselectAllItems() {
  document.querySelectorAll('.canvas-item.selected').forEach(el => {
    el.classList.remove('selected');
  });
  selectedCanvasItem = null;
}

function bringForward(element) {
  const current = parseInt(element.style.zIndex) || 0;
  element.style.zIndex = current + 1;
}

function sendBackward(element) {
  const current = parseInt(element.style.zIndex) || 0;
  element.style.zIndex = Math.max(0, current - 1);
}

// Drag handling for canvas items
function setupCanvasItemDrag(element) {
  let isDragging = false;
  let startX, startY, initialX, initialY;

  element.addEventListener('mousedown', startDrag);
  element.addEventListener('touchstart', startDrag, { passive: false });

  function startDrag(e) {
    if (e.target.tagName === 'BUTTON') return;

    isDragging = true;
    selectCanvasItem(element);

    const touch = e.touches ? e.touches[0] : e;
    startX = touch.clientX;
    startY = touch.clientY;
    initialX = element.offsetLeft;
    initialY = element.offsetTop;

    document.addEventListener('mousemove', drag);
    document.addEventListener('mouseup', stopDrag);
    document.addEventListener('touchmove', drag, { passive: false });
    document.addEventListener('touchend', stopDrag);

    e.preventDefault();
  }

  function drag(e) {
    if (!isDragging) return;

    const touch = e.touches ? e.touches[0] : e;
    const canvas = document.querySelector('.outfit-canvas');
    const canvasRect = canvas.getBoundingClientRect();

    let newX = initialX + (touch.clientX - startX);
    let newY = initialY + (touch.clientY - startY);

    // Constrain to canvas
    newX = Math.max(0, Math.min(newX, canvasRect.width - element.offsetWidth));
    newY = Math.max(0, Math.min(newY, canvasRect.height - element.offsetHeight));

    element.style.left = `${newX}px`;
    element.style.top = `${newY}px`;

    e.preventDefault();
  }

  function stopDrag() {
    isDragging = false;
    document.removeEventListener('mousemove', drag);
    document.removeEventListener('mouseup', stopDrag);
    document.removeEventListener('touchmove', drag);
    document.removeEventListener('touchend', stopDrag);

    // Update position in currentOutfitItems
    const itemId = parseInt(element.dataset.itemId);
    const item = currentOutfitItems.find(i => i.id === itemId);
    if (item) {
      item.x = element.offsetLeft;
      item.y = element.offsetTop;
      item.z = parseInt(element.style.zIndex) || 0;
    }
  }
}

function clearCanvas() {
  const canvas = document.querySelector('.outfit-canvas');
  if (!canvas) return;

  canvas.querySelectorAll('.canvas-item').forEach(el => el.remove());
  currentOutfitItems = [];
  selectedCanvasItem = null;

  renderPickerItems(document.querySelector('.picker-tab.active')?.dataset.category || 'all');
  showToast('Canvas cleared', 'info');
}

async function handleDressWithAI() {
  if (!aiAvailable) {
    showToast('AI not available. Start the API server.', 'error');
    return;
  }

  const canvas = document.querySelector('.outfit-canvas');
  const bodyBg = canvas?.style.backgroundImage;

  if (!bodyBg || bodyBg === 'none') {
    showToast('Please upload a body/mannequin image first!', 'error');
    return;
  }

  if (currentOutfitItems.length === 0) {
    showToast('Please add clothing items to the canvas first!', 'error');
    return;
  }

  // Extract the data URL from backgroundImage
  const match = bodyBg.match(/url\("?(.+?)"?\)/);
  if (!match) {
    showToast('Could not read body image', 'error');
    return;
  }
  const bodyDataUrl = match[1];

  // Prepare clothing items data
  const clothingData = currentOutfitItems.map(item => ({
    id: item.id,
    name: item.name,
    category: item.category,
    image: item.image
  }));

  showProcessingModal('AI is positioning clothing on body...');

  try {
    const result = await dressWithAI(bodyDataUrl, clothingData);

    if (result && result.positions) {
      console.log('AI Dress result:', result);

      // Apply the suggested positions
      const canvasRect = canvas.getBoundingClientRect();

      result.positions.forEach(pos => {
        const canvasItem = document.querySelector(`.canvas-item[data-item-id="${pos.id}"]`);
        if (canvasItem) {
          // Convert percentage to pixels
          const x = (pos.x / 100) * canvasRect.width;
          const y = (pos.y / 100) * canvasRect.height;

          canvasItem.style.left = `${x}px`;
          canvasItem.style.top = `${y}px`;
          canvasItem.style.zIndex = pos.z;

          // Update in currentOutfitItems
          const item = currentOutfitItems.find(i => i.id === pos.id);
          if (item) {
            item.x = x;
            item.y = y;
            item.z = pos.z;
          }
        }
      });

      closeModal();
      showToast(result.explanation || 'AI positioned clothing on body!', 'success');
    } else {
      closeModal();
      showToast('AI could not position clothing', 'error');
    }
  } catch (e) {
    closeModal();
    showToast('AI dress failed: ' + e.message, 'error');
  }
}

async function handleRateOutfit() {
  if (!aiAvailable) {
    showToast('AI not available. Start the API server.', 'error');
    return;
  }

  if (currentOutfitItems.length === 0) {
    showToast('Add items to your outfit first!', 'error');
    return;
  }

  // Prepare items for rating
  const items = currentOutfitItems.map(item => ({
    id: item.id,
    name: item.name,
    category: item.category,
    color: item.color,
    tags: item.tags || []
  }));

  showProcessingModal('AI is rating your outfit...');

  try {
    const result = await rateOutfitWithAI(items);

    if (result) {
      closeModal();

      const ratingStars = '⭐'.repeat(Math.round(result.rating / 2));
      const strengthsList = result.strengths && result.strengths.length > 0 ?
        `<ul>${result.strengths.map(s => `<li>${s}</li>`).join('')}</ul>` : '';
      const improvementsList = result.improvements && result.improvements.length > 0 ?
        `<ul>${result.improvements.map(i => `<li>${i}</li>`).join('')}</ul>` : '';

      showModal('⭐ Outfit Rating', `
        <div style="text-align: center; margin: 1.5rem 0;">
          <div style="font-size: 3rem;">${ratingStars}</div>
          <div style="font-size: 2rem; font-weight: bold; color: var(--primary);">${result.rating}/10</div>
        </div>
        <div style="padding: 1rem; background: var(--warm-bg); border-radius: 0.5rem; margin-bottom: 1rem;">
          <p style="margin: 0; color: var(--text-neutral);">${result.feedback}</p>
        </div>
        ${result.strengths && result.strengths.length > 0 ? `
          <div style="margin-bottom: 1rem;">
            <strong style="color: #27AE60;">✅ Strengths:</strong>
            ${strengthsList}
          </div>
        ` : ''}
        ${result.improvements && result.improvements.length > 0 ? `
          <div>
            <strong style="color: var(--primary);">💡 How to Improve:</strong>
            ${improvementsList}
          </div>
        ` : ''}
      `, [
        { label: 'Close', onclick: 'closeModal()' }
      ]);
    } else {
      closeModal();
      showToast('AI could not rate outfit', 'error');
    }
  } catch (e) {
    closeModal();
    showToast('Outfit rating failed: ' + e.message, 'error');
  }
}

function saveOutfit() {
  if (currentOutfitItems.length === 0) {
    showToast('Add items to your outfit first!', 'error');
    return;
  }

  showModal('Save Outfit', `
    <div class="form-group">
      <label>Outfit Name</label>
      <input type="text" id="outfit-name-input" placeholder="e.g., Casual Friday" autofocus>
    </div>
    <div class="form-group">
      <label>Notes (optional)</label>
      <textarea id="outfit-notes-input" rows="2" placeholder="Great for warm weather..."></textarea>
    </div>
  `, [
    { label: 'Cancel', onclick: 'closeModal()' },
    { label: 'Save Outfit', primary: true, onclick: 'confirmSaveOutfit()' },
  ]);
}

function confirmSaveOutfit() {
  const nameInput = document.getElementById('outfit-name-input');
  const notesInput = document.getElementById('outfit-notes-input');

  const name = nameInput?.value.trim();
  if (!name) {
    showToast('Please enter a name for your outfit', 'error');
    return;
  }

  // Get body background if set
  const canvas = document.querySelector('.outfit-canvas');
  const bodyBackground = canvas?.style.backgroundImage !== 'none' ? canvas?.style.backgroundImage : null;

  const newOutfit = {
    id: Date.now(),
    name: name,
    notes: notesInput?.value.trim() || '',
    itemIds: currentOutfitItems.map(i => i.id),
    positions: currentOutfitItems.map(i => ({ id: i.id, x: i.x, y: i.y, z: i.z })),
    bodyBackground: bodyBackground, // Save the body/mannequin image
    wearCount: 0,
    rating: 0,
    createdAt: new Date().toISOString(),
  };

  appData.outfits.push(newOutfit);
  saveData(appData);

  closeModal();
  showToast(`Outfit "${name}" saved!`, 'success');

  setTimeout(() => {
    window.location.href = 'outfits.html';
  }, 1000);
}

// ============================================================================
// ITEM DETAIL PAGE
// ============================================================================

function initItemDetail() {
  const container = document.getElementById('item-detail');
  if (!container) return;

  const params = new URLSearchParams(window.location.search);
  const itemId = parseInt(params.get('id'));

  const item = appData.items.find(i => i.id === itemId);
  if (!item) {
    showToast('Item not found', 'error');
    return;
  }

  // Populate form
  document.getElementById('item-name').value = item.name || '';
  document.getElementById('item-category').value = item.category || 'Uncategorized';
  document.getElementById('item-price').value = item.price || 0;
  document.getElementById('item-brand').value = item.brand || '';
  document.getElementById('item-color').value = item.color || '';
  document.getElementById('item-image-preview').src = item.image;

  // Show tags if present
  if (item.tags && item.tags.length > 0) {
    const tagsSection = document.getElementById('tags-section');
    const tagsContainer = document.getElementById('item-tags');
    if (tagsSection && tagsContainer) {
      tagsSection.style.display = 'block';
      renderTags(tagsContainer, item.tags);
    }
  }

  // Show AI analysis if present
  if (item.aiAnalysis) {
    const analysisSection = document.getElementById('ai-analysis-section');
    const analysisText = document.getElementById('ai-analysis-text');
    if (analysisSection && analysisText) {
      analysisSection.style.display = 'block';
      analysisText.textContent = item.aiAnalysis;
    }
  }

  // Setup enhance brand button (uses 13b model)
  const enhanceBrandBtn = document.getElementById('enhance-brand-btn');
  if (enhanceBrandBtn) {
    enhanceBrandBtn.onclick = async () => {
      if (!aiAvailable) {
        showToast('AI not available. Start the API server.', 'error');
        return;
      }

      enhanceBrandBtn.disabled = true;
      enhanceBrandBtn.textContent = '⏳ Analyzing (2-3 min)...';

      // Use llava:13b for better brand detection
      const aiResult = await analyzeImageWithAI(item.image, 'llava:13b');

      if (aiResult && aiResult.brand) {
        document.getElementById('item-brand').value = aiResult.brand;
        showToast(`Brand detected: ${aiResult.brand}`, 'success');

        // Auto-save the brand
        item.brand = aiResult.brand;
        saveData(appData);
      } else {
        showToast('Could not detect brand. Try manually entering it.', 'error');
      }

      enhanceBrandBtn.disabled = false;
      enhanceBrandBtn.textContent = '✨ Detect with AI (13b)';
    };
  }

  // Setup regenerate tags button
  const regenBtn = document.getElementById('regenerate-tags-btn');
  if (regenBtn) {
    regenBtn.onclick = async () => {
      if (!aiAvailable) {
        showToast('AI not available. Start the API server.', 'error');
        return;
      }
      regenBtn.disabled = true;
      regenBtn.textContent = '⏳ Generating...';

      const description = `${item.name} - ${item.category} - ${item.color}`;
      const newTags = await generateTagsWithAI(description);

      if (newTags.length > 0) {
        item.tags = newTags;
        saveData(appData);

        const tagsContainer = document.getElementById('item-tags');
        if (tagsContainer) {
          renderTags(tagsContainer, newTags);
        }
        showToast('Tags regenerated!', 'success');
      } else {
        showToast('Failed to generate tags', 'error');
      }

      regenBtn.disabled = false;
      regenBtn.textContent = '🔄 Regenerate with AI';
    };
  }

  // Setup save button
  const saveBtn = document.getElementById('save-item-btn');
  if (saveBtn) {
    saveBtn.onclick = () => saveItem(itemId);
  }

  // Setup delete button
  const deleteBtn = document.getElementById('delete-item-btn');
  if (deleteBtn) {
    deleteBtn.onclick = () => deleteItem(itemId);
  }

  // Setup wear button
  const wearBtn = document.getElementById('wear-item-btn');
  if (wearBtn) {
    wearBtn.onclick = () => logWear(itemId);
  }

  // Find Matching Items button
  const findMatchesBtn = document.getElementById('find-matches-btn');
  if (findMatchesBtn) {
    findMatchesBtn.onclick = async () => {
      if (!aiAvailable) {
        showToast('AI not available. Start the API server.', 'error');
        return;
      }

      const item = appData.items.find(i => i.id === itemId);
      if (!item) return;

      const baseItem = {
        id: item.id,
        name: item.name,
        category: item.category,
        color: item.color,
        brand: item.brand,
        tags: item.tags || []
      };

      const otherItems = appData.items
        .filter(i => i.id !== itemId)
        .map(i => ({
          id: i.id,
          name: i.name,
          category: i.category,
          color: i.color,
          brand: i.brand,
          tags: i.tags || []
        }));

      if (otherItems.length === 0) {
        showToast('No other items in wardrobe to match with!', 'error');
        return;
      }

      findMatchesBtn.disabled = true;
      findMatchesBtn.textContent = '⏳ Finding matches...';

      const result = await findStyleMatches(baseItem, otherItems, 6);

      findMatchesBtn.disabled = false;
      findMatchesBtn.textContent = '✨ Find Matching Items';

      if (result && result.matching_items && result.matching_items.length > 0) {
        displayStyleMatches(result.matching_items);
      } else {
        showToast('Could not find matching items', 'error');
      }
    };
  }
}

function displayStyleMatches(matchingItems) {
  const section = document.getElementById('style-matches-section');
  const grid = document.getElementById('style-matches-grid');

  if (!section || !grid) return;

  grid.innerHTML = '';

  matchingItems.forEach(match => {
    const item = appData.items.find(i => i.id === match.id);
    if (!item) return;

    const card = document.createElement('div');
    card.className = 'item-card';
    card.onclick = () => window.location.href = `item-detail.html?id=${item.id}`;

    const scoreColor = match.score >= 80 ? '#27AE60' :
                       match.score >= 60 ? '#F39C12' : '#E74C3C';

    card.innerHTML = `
      <img src="${item.image}" alt="${item.name}" class="item-image" loading="lazy">
      <div class="item-info">
        <div class="item-name">${item.name}</div>
        <div class="item-meta">${item.category} • ${item.color}</div>
        <div style="margin-top: 0.5rem; padding: 0.5rem; background: var(--warm-bg); border-radius: 0.25rem;">
          <div style="font-size: 0.75rem; color: var(--text-neutral); margin-bottom: 0.25rem;">
            Match Score: <strong style="color: ${scoreColor};">${match.score}/100</strong>
          </div>
          <div style="font-size: 0.7rem; color: var(--text-neutral); line-height: 1.4;">
            ${match.reason}
          </div>
        </div>
      </div>
    `;

    grid.appendChild(card);
  });

  section.style.display = 'block';
  section.scrollIntoView({ behavior: 'smooth', block: 'start' });
  showToast('Found matching items!', 'success');
}

function renderTags(container, tags) {
  container.innerHTML = tags.map(tag =>
    `<span style="background: var(--primary-light); color: var(--text-primary); padding: 0.25rem 0.75rem; border-radius: 1rem; font-size: 0.875rem;">${tag}</span>`
  ).join('');
}

function saveItem(itemId) {
  const item = appData.items.find(i => i.id === itemId);
  if (!item) return;

  item.name = document.getElementById('item-name').value;
  item.category = document.getElementById('item-category').value;
  item.price = parseFloat(document.getElementById('item-price').value) || 0;
  item.brand = document.getElementById('item-brand').value;
  item.color = document.getElementById('item-color').value;

  saveData(appData);
  showToast('Item saved!', 'success');

  setTimeout(() => {
    window.location.href = 'wardrobe.html';
  }, 1000);
}

function deleteItem(itemId) {
  if (!confirm('Are you sure you want to delete this item? This cannot be undone.')) {
    return;
  }

  appData.items = appData.items.filter(i => i.id !== itemId);

  // Also remove from outfits
  appData.outfits.forEach(outfit => {
    outfit.itemIds = outfit.itemIds.filter(id => id !== itemId);
    outfit.positions = outfit.positions.filter(p => p.id !== itemId);
  });

  saveData(appData);
  showToast('Item deleted', 'success');

  setTimeout(() => {
    window.location.href = 'wardrobe.html';
  }, 500);
}

function logWear(itemId) {
  const item = appData.items.find(i => i.id === itemId);
  if (!item) return;

  item.wearCount = (item.wearCount || 0) + 1;
  appData.wearHistory.push({
    itemId: itemId,
    date: new Date().toISOString(),
  });

  saveData(appData);

  const costPerWear = (item.price / item.wearCount).toFixed(2);
  showToast(`Logged wear! Cost per wear is now $${costPerWear}`, 'success');
}

// ============================================================================
// OUTFITS LIST
// ============================================================================

function renderOutfits(containerId) {
  const container = document.getElementById(containerId);
  if (!container) return;

  container.innerHTML = '';

  appData.outfits.forEach(outfit => {
    const items = outfit.itemIds.map(id => appData.items.find(i => i.id === id)).filter(Boolean);

    const card = document.createElement('div');
    card.className = 'outfit-card';
    card.onclick = () => editOutfit(outfit.id);

    card.innerHTML = `
      <div class="outfit-preview">
        ${items.slice(0, 3).map((item, i) => `
          <img src="${item.image}" alt="${item.name}" style="position: absolute; left: ${50 + i * 30}px; top: ${30 + i * 40}px; max-width: 100px; max-height: 100px;">
        `).join('')}
      </div>
      <div class="outfit-info">
        <h3>${outfit.name}</h3>
        <p class="outfit-meta">${items.length} items • Worn ${outfit.wearCount} times</p>
        <div class="outfit-rating">
          ${'★'.repeat(Math.floor(outfit.rating || 0))}${'☆'.repeat(5 - Math.floor(outfit.rating || 0))}
          <span>${outfit.rating || 'Not rated'}</span>
        </div>
      </div>
    `;

    container.appendChild(card);
  });

  // Add "create new" card
  const createCard = document.createElement('div');
  createCard.className = 'outfit-card create-outfit-card';
  createCard.onclick = () => window.location.href = 'outfit-builder.html';
  createCard.innerHTML = `
    <div class="create-outfit-icon">+</div>
    <p>Create New Outfit</p>
  `;
  container.appendChild(createCard);
}

function editOutfit(outfitId) {
  // Load outfit into builder
  const outfit = appData.outfits.find(o => o.id === outfitId);
  if (!outfit) return;

  sessionStorage.setItem('editOutfitId', outfitId.toString());
  window.location.href = 'outfit-builder.html';
}

// ============================================================================
// ANALYTICS
// ============================================================================

function renderAnalytics() {
  // Update stat cards
  const totalItems = appData.items.length;
  const totalInvestment = appData.items.reduce((sum, item) => sum + (item.price || 0), 0);
  const totalWears = appData.items.reduce((sum, item) => sum + (item.wearCount || 0), 0);
  const avgCostPerWear = totalWears > 0 ? (totalInvestment / totalWears) : 0;

  updateStatCard('stat-total-items', totalItems);
  updateStatCard('stat-total-investment', `$${totalInvestment.toFixed(0)}`);
  updateStatCard('stat-avg-cpw', `$${avgCostPerWear.toFixed(2)}`);
  updateStatCard('stat-total-wears', totalWears);
}

function updateStatCard(id, value) {
  const el = document.getElementById(id);
  if (el) el.textContent = value;
}

// ============================================================================
// INITIALIZATION
// ============================================================================

document.addEventListener('DOMContentLoaded', function() {
  // Check AI status on load
  checkAIStatus();

  // Upload page
  initUpload();

  // Wardrobe page
  renderWardrobe('wardrobe-grid');
  initWardrobeFilter();

  // Outfit builder
  initOutfitBuilder();

  // Item detail
  initItemDetail();

  // Outfits list
  renderOutfits('outfits-grid');

  // Analytics
  renderAnalytics();

  // Load outfit if editing
  const editOutfitId = sessionStorage.getItem('editOutfitId');
  if (editOutfitId && document.querySelector('.outfit-canvas')) {
    const outfit = appData.outfits.find(o => o.id === parseInt(editOutfitId));
    if (outfit) {
      // Restore body background if it exists
      const canvas = document.querySelector('.outfit-canvas');
      if (outfit.bodyBackground && canvas) {
        canvas.style.backgroundImage = outfit.bodyBackground;
      }

      // Load items
      outfit.itemIds.forEach(itemId => {
        const item = appData.items.find(i => i.id === itemId);
        if (item) {
          const pos = outfit.positions.find(p => p.id === itemId);
          addToCanvasAtPosition(item, pos?.x || 100, pos?.y || 100, pos?.z || 0);
        }
      });
      sessionStorage.removeItem('editOutfitId');
    }
  }
});

function addToCanvasAtPosition(item, x, y, z) {
  const canvas = document.querySelector('.outfit-canvas');
  if (!canvas) return;

  const canvasItem = document.createElement('div');
  canvasItem.className = 'canvas-item';
  canvasItem.dataset.itemId = item.id;
  canvasItem.style.left = `${x}px`;
  canvasItem.style.top = `${y}px`;
  canvasItem.style.zIndex = z;

  canvasItem.innerHTML = `
    <img src="${item.image}" alt="${item.name}" draggable="false">
    <div class="canvas-item-controls">
      <button onclick="bringForward(this.closest('.canvas-item'))" title="Bring forward">↑</button>
      <button onclick="sendBackward(this.closest('.canvas-item'))" title="Send backward">↓</button>
      <button onclick="removeFromCanvas(${item.id})" title="Remove">×</button>
    </div>
  `;

  setupCanvasItemDrag(canvasItem);
  canvasItem.addEventListener('click', (e) => {
    if (e.target.tagName !== 'BUTTON') {
      selectCanvasItem(canvasItem);
    }
  });

  canvas.appendChild(canvasItem);
  currentOutfitItems.push({ ...item, x, y, z });
}

// Export functions for HTML onclick handlers
window.showToast = showToast;
window.showModal = showModal;
window.closeModal = closeModal;
window.clearCanvas = clearCanvas;
window.saveOutfit = saveOutfit;
window.confirmSaveOutfit = confirmSaveOutfit;
window.addToCanvas = addToCanvas;
window.removeFromCanvas = removeFromCanvas;
window.bringForward = bringForward;
window.sendBackward = sendBackward;
window.logWear = logWear;
window.initAIStylist = initAIStylist;

// ============================================================================
// AI STYLIST PAGE
// ============================================================================

let selectedColorMatchItems = [];

function initAIStylist() {
  // Suggest Outfit
  const suggestBtn = document.getElementById('suggest-outfit-btn');
  if (suggestBtn) {
    suggestBtn.onclick = async () => {
      if (!aiAvailable) {
        showToast('AI is not available. Start the mockup server first.', 'error');
        return;
      }

      const occasion = document.getElementById('occasion-select').value;
      const weather = document.getElementById('weather-select').value;

      // Convert wardrobe items to simple format
      const items = appData.items.map(item => ({
        id: item.id,
        name: item.name,
        category: item.category,
        color: item.color,
        brand: item.brand,
        tags: item.tags || []
      }));

      if (items.length === 0) {
        showToast('Your wardrobe is empty! Upload some items first.', 'error');
        return;
      }

      suggestBtn.disabled = true;
      suggestBtn.textContent = '⏳ Thinking...';

      const result = await suggestOutfitWithAI(occasion, weather, items);

      suggestBtn.disabled = false;
      suggestBtn.textContent = '✨ Suggest Outfit';

      if (result) {
        displayOutfitSuggestion(result, items);
      }
    };
  }

  // Color Match
  renderColorMatchItemSelector();

  const analyzeColorsBtn = document.getElementById('analyze-colors-btn');
  if (analyzeColorsBtn) {
    analyzeColorsBtn.onclick = async () => {
      if (selectedColorMatchItems.length < 2) {
        showToast('Select at least 2 items to analyze colors', 'error');
        return;
      }

      analyzeColorsBtn.disabled = true;
      analyzeColorsBtn.textContent = '⏳ Analyzing...';

      const items = selectedColorMatchItems.map(id => {
        const item = appData.items.find(i => i.id === id);
        return {
          id: item.id,
          name: item.name,
          category: item.category,
          color: item.color,
          tags: item.tags || []
        };
      });

      const result = await analyzeColorMatch(items);

      analyzeColorsBtn.disabled = false;
      analyzeColorsBtn.textContent = '🎨 Analyze Colors';

      if (result) {
        displayColorMatchResult(result);
      }
    };
  }

  // Wardrobe Analysis
  const analyzeWardrobeBtn = document.getElementById('analyze-wardrobe-btn');
  if (analyzeWardrobeBtn) {
    analyzeWardrobeBtn.onclick = async () => {
      if (!aiAvailable) {
        showToast('AI is not available. Start the mockup server first.', 'error');
        return;
      }

      const items = appData.items.map(item => ({
        id: item.id,
        name: item.name,
        category: item.category,
        color: item.color,
        brand: item.brand,
        tags: item.tags || []
      }));

      if (items.length === 0) {
        showToast('Your wardrobe is empty! Upload some items first.', 'error');
        return;
      }

      analyzeWardrobeBtn.disabled = true;
      analyzeWardrobeBtn.textContent = '⏳ Analyzing...';

      const result = await analyzeWardrobeGaps(items);

      analyzeWardrobeBtn.disabled = false;
      analyzeWardrobeBtn.textContent = '🔍 Analyze My Wardrobe';

      if (result) {
        displayWardrobeAnalysisResult(result);
      }
    };
  }
}

function renderColorMatchItemSelector() {
  const container = document.getElementById('color-match-items');
  if (!container) return;

  container.innerHTML = '';

  if (appData.items.length === 0) {
    container.innerHTML = '<p style="color: var(--text-neutral); text-align: center;">No items in wardrobe</p>';
    return;
  }

  appData.items.forEach(item => {
    const itemEl = document.createElement('div');
    itemEl.className = 'item-selector-card';
    itemEl.style.cssText = 'display: inline-block; margin: 0.5rem; padding: 0.5rem; border: 2px solid var(--border); border-radius: 0.5rem; cursor: pointer; transition: all 0.2s;';

    itemEl.onclick = () => toggleColorMatchItem(item.id, itemEl);

    itemEl.innerHTML = `
      <img src="${item.image}" alt="${item.name}" style="width: 80px; height: 80px; object-fit: cover; border-radius: 0.25rem;">
      <div style="font-size: 0.75rem; margin-top: 0.25rem; text-align: center;">
        ${item.name.substring(0, 15)}${item.name.length > 15 ? '...' : ''}
      </div>
      <div style="font-size: 0.7rem; color: var(--text-neutral); text-align: center;">
        ${item.color}
      </div>
    `;

    container.appendChild(itemEl);
  });
}

function toggleColorMatchItem(itemId, element) {
  const index = selectedColorMatchItems.indexOf(itemId);

  if (index > -1) {
    selectedColorMatchItems.splice(index, 1);
    element.style.borderColor = 'var(--border)';
    element.style.background = 'transparent';
  } else {
    selectedColorMatchItems.push(itemId);
    element.style.borderColor = 'var(--primary)';
    element.style.background = 'var(--warm-bg)';
  }

  // Enable/disable analyze button
  const analyzeBtn = document.getElementById('analyze-colors-btn');
  if (analyzeBtn) {
    analyzeBtn.disabled = selectedColorMatchItems.length < 2;
  }
}

function displayOutfitSuggestion(result, allItems) {
  const container = document.getElementById('outfit-suggestion-result');
  if (!container) return;

  const selectedItems = result.selected_items.map(id => allItems.find(i => i.id === id)).filter(Boolean);

  let itemsHTML = '';
  selectedItems.forEach(item => {
    const fullItem = appData.items.find(i => i.id === item.id);
    if (fullItem) {
      itemsHTML += `
        <div style="display: inline-block; margin: 0.5rem; text-align: center;">
          <img src="${fullItem.image}" alt="${item.name}" style="width: 100px; height: 100px; object-fit: cover; border-radius: 0.5rem; border: 2px solid var(--primary);">
          <div style="font-size: 0.75rem; margin-top: 0.5rem;">${item.name}</div>
          <div style="font-size: 0.7rem; color: var(--text-neutral);">${item.category} • ${item.color}</div>
        </div>
      `;
    }
  });

  let tipsHTML = '';
  if (result.tips && result.tips.length > 0) {
    tipsHTML = `
      <div style="margin-top: 1rem; padding-top: 1rem; border-top: 1px solid var(--border);">
        <strong>Styling Tips:</strong>
        <ul style="margin: 0.5rem 0; padding-left: 1.5rem;">
          ${result.tips.map(tip => `<li>${tip}</li>`).join('')}
        </ul>
      </div>
    `;
  }

  container.innerHTML = `
    <h3 style="margin-top: 0; color: var(--primary);">✨ Suggested Outfit</h3>
    <div style="margin: 1rem 0;">
      ${itemsHTML}
    </div>
    <div style="padding: 1rem; background: white; border-radius: 0.5rem; margin-top: 1rem;">
      <strong>Why this works:</strong>
      <p style="margin: 0.5rem 0 0 0; color: var(--text-neutral);">${result.reasoning}</p>
    </div>
    ${tipsHTML}
    <button class="btn btn-primary" style="margin-top: 1rem; width: 100%;" onclick="createOutfitFromSuggestion(${JSON.stringify(result.selected_items).replace(/"/g, '&quot;')})">
      📸 Create Outfit on Canvas
    </button>
  `;

  container.style.display = 'block';
  showToast('Outfit suggested! Check the results below.', 'success');
}

function displayColorMatchResult(result) {
  const container = document.getElementById('color-match-result');
  if (!container) return;

  const scoreColor = result.compatibility_score >= 80 ? '#27AE60' :
                     result.compatibility_score >= 60 ? '#F39C12' : '#E74C3C';

  let suggestionsHTML = '';
  if (result.suggestions && result.suggestions.length > 0) {
    suggestionsHTML = `
      <div style="margin-top: 1rem; padding-top: 1rem; border-top: 1px solid var(--border);">
        <strong>Suggestions:</strong>
        <ul style="margin: 0.5rem 0; padding-left: 1.5rem;">
          ${result.suggestions.map(s => `<li>${s}</li>`).join('')}
        </ul>
      </div>
    `;
  }

  container.innerHTML = `
    <h3 style="margin-top: 0; color: var(--primary);">🎨 Color Analysis</h3>
    <div style="text-align: center; margin: 1.5rem 0;">
      <div style="font-size: 3rem; font-weight: bold; color: ${scoreColor};">
        ${result.compatibility_score}
      </div>
      <div style="color: var(--text-neutral); font-size: 0.875rem;">
        Compatibility Score
      </div>
    </div>
    <div style="padding: 1rem; background: white; border-radius: 0.5rem;">
      <p style="margin: 0; color: var(--text-neutral);">${result.analysis}</p>
    </div>
    ${suggestionsHTML}
  `;

  container.style.display = 'block';
  showToast('Color analysis complete!', 'success');
}

function displayWardrobeAnalysisResult(result) {
  const container = document.getElementById('wardrobe-analysis-result');
  if (!container) return;

  let missingItemsHTML = '';
  if (result.missing_items && result.missing_items.length > 0) {
    missingItemsHTML = `
      <div class="card">
        <h3 style="color: var(--primary);">🛍️ Suggested Items to Add</h3>
        <ul style="line-height: 2;">
          ${result.missing_items.map(item => `<li>${item}</li>`).join('')}
        </ul>
      </div>
    `;
  }

  let strengthsHTML = '';
  if (result.strengths && result.strengths.length > 0) {
    strengthsHTML = `
      <div class="card">
        <h3 style="color: #27AE60;">✅ What You Have Covered</h3>
        <ul style="line-height: 2;">
          ${result.strengths.map(s => `<li>${s}</li>`).join('')}
        </ul>
      </div>
    `;
  }

  let tipsHTML = '';
  if (result.tips && result.tips.length > 0) {
    tipsHTML = `
      <div class="card" style="background: var(--warm-bg);">
        <h3>💡 Wardrobe Tips</h3>
        <ul style="line-height: 2;">
          ${result.tips.map(tip => `<li>${tip}</li>`).join('')}
        </ul>
      </div>
    `;
  }

  container.innerHTML = `
    <div class="card">
      <h3 style="margin-top: 0;">📊 Overall Assessment</h3>
      <p style="color: var(--text-neutral); line-height: 1.6;">${result.summary}</p>
    </div>
    ${missingItemsHTML}
    ${strengthsHTML}
    ${tipsHTML}
  `;

  container.style.display = 'block';
  showToast('Wardrobe analysis complete!', 'success');
}

window.createOutfitFromSuggestion = function(itemIds) {
  // Save selected items to session storage and redirect
  sessionStorage.setItem('suggested_outfit_items', JSON.stringify(itemIds));
  window.location.href = 'outfit-builder.html?suggested=true';
};
