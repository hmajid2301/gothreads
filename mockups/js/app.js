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

// ============================================================================
// GENERIC AI JOB SUBMISSION + SSE + NOTIFICATION PANEL
// ============================================================================

let _notifCounter = 0;
window._activeJobs = {};

function initNotificationPanel() {
  if (document.getElementById('ai-job-notifications')) return;
  const panel = document.createElement('div');
  panel.id = 'ai-job-notifications';
  panel.style.cssText = `
    position: fixed;
    bottom: 5rem;
    right: 1.5rem;
    width: 320px;
    max-height: 60vh;
    overflow-y: auto;
    z-index: 1100;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    pointer-events: none;
  `;
  document.body.appendChild(panel);
}

function addJobNotification(jobId, label) {
  const notifId = `notif-${++_notifCounter}`;
  const panel = document.getElementById('ai-job-notifications');
  if (!panel) return notifId;

  const card = document.createElement('div');
  card.id = notifId;
  card.dataset.jobId = jobId;
  card.style.cssText = `
    background: white;
    border-radius: 0.75rem;
    padding: 0.75rem 1rem;
    box-shadow: 0 4px 12px rgba(0,0,0,0.15);
    pointer-events: auto;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    border-left: 4px solid var(--primary, #6C5CE7);
    animation: notifSlideIn 0.3s ease;
  `;
  card.innerHTML = `
    <div class="notif-spinner" style="width:20px;height:20px;border:3px solid #eee;border-top-color:var(--primary, #6C5CE7);border-radius:50%;animation:spin 0.8s linear infinite;flex-shrink:0;"></div>
    <div style="flex:1;min-width:0;">
      <div style="font-weight:600;font-size:0.85rem;color:var(--text-primary, #2d3436);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;">${label}</div>
      <div class="notif-status" style="font-size:0.75rem;color:var(--text-neutral, #95a5a6);margin-top:2px;">Queued...</div>
    </div>
    <button onclick="cancelAIJob('${jobId}')" style="background:none;border:none;cursor:pointer;color:var(--text-neutral, #95a5a6);font-size:1.2rem;padding:0 4px;line-height:1;" title="Cancel">&times;</button>
  `;
  panel.appendChild(card);
  return notifId;
}

function updateJobNotification(notifId, status, message) {
  const card = document.getElementById(notifId);
  if (!card) return;

  const statusEl = card.querySelector('.notif-status');
  if (statusEl) statusEl.textContent = message;

  const spinner = card.querySelector('.notif-spinner');
  const cancelBtn = card.querySelector('button');

  if (status === 'complete') {
    card.style.borderLeftColor = '#27AE60';
    if (spinner) spinner.outerHTML = '<span style="font-size:1.2rem;color:#27AE60;flex-shrink:0;">&#10003;</span>';
    if (cancelBtn) cancelBtn.style.display = 'none';
  } else if (status === 'failed' || status === 'cancelled') {
    card.style.borderLeftColor = '#E74C3C';
    if (spinner) spinner.outerHTML = '<span style="font-size:1.2rem;color:#E74C3C;flex-shrink:0;">&#10007;</span>';
    if (cancelBtn) cancelBtn.style.display = 'none';
  }
}

function autoDismissNotification(notifId, delayMs = 4000) {
  setTimeout(() => {
    const card = document.getElementById(notifId);
    if (card) {
      card.style.transition = 'opacity 0.3s, transform 0.3s';
      card.style.opacity = '0';
      card.style.transform = 'translateX(100px)';
      setTimeout(() => card.remove(), 300);
    }
  }, delayMs);
}

function cancelAIJob(jobId) {
  if (window._activeJobs && window._activeJobs[jobId]) {
    window._activeJobs[jobId].cancel();
    delete window._activeJobs[jobId];
  }
}

/**
 * Submit an AI job and return a Promise that resolves with the result.
 * Shows progress in the bottom-right notification panel via SSE.
 */
async function submitAIJob(type, payload, label = null) {
  const displayLabel = label || type.replace(/-/g, ' ');

  const resp = await fetch(`${AI_API_URL}/ai/jobs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ type, payload })
  });

  if (!resp.ok) {
    const err = await resp.json().catch(() => ({ error: 'Submit failed' }));
    throw new Error(err.error || `Submit failed (${resp.status})`);
  }

  const { id: jobId } = await resp.json();
  const notifId = addJobNotification(jobId, displayLabel);

  return new Promise((resolve, reject) => {
    const evtSource = new EventSource(`${AI_API_URL}/ai/jobs/${jobId}`);
    let settled = false;

    evtSource.addEventListener('status', (e) => {
      updateJobNotification(notifId, 'processing', e.data);
    });

    evtSource.addEventListener('result', (e) => {
      evtSource.close();
      if (!settled) {
        settled = true;
        updateJobNotification(notifId, 'complete', 'Done!');
        autoDismissNotification(notifId, 4000);
        resolve(JSON.parse(e.data));
      }
    });

    evtSource.addEventListener('error', (e) => {
      evtSource.close();
      if (!settled) {
        settled = true;
        const msg = e.data || 'Failed';
        updateJobNotification(notifId, 'failed', msg);
        autoDismissNotification(notifId, 6000);
        reject(new Error(msg));
      }
    });

    evtSource.onerror = () => {
      evtSource.close();
      if (!settled) {
        settled = true;
        updateJobNotification(notifId, 'failed', 'Lost connection');
        autoDismissNotification(notifId, 6000);
        reject(new Error('Lost connection to server'));
      }
    };

    window._activeJobs[jobId] = {
      cancel: () => {
        evtSource.close();
        fetch(`${AI_API_URL}/ai/jobs/${jobId}`, { method: 'DELETE' });
        if (!settled) {
          settled = true;
          updateJobNotification(notifId, 'cancelled', 'Cancelled');
          autoDismissNotification(notifId, 3000);
          reject(new Error('Cancelled'));
        }
      }
    };
  });
}

// ============================================================================
// AI WRAPPER FUNCTIONS (delegate to submitAIJob)
// ============================================================================

async function analyzeImageWithAI(imageDataUrl, model = null, customPrompt = null) {
  if (!aiAvailable) return null;
  try {
    const payload = { image: imageDataUrl };
    if (model) payload.model = model;
    if (customPrompt) payload.prompt = customPrompt;
    return await submitAIJob('analyze', payload, 'Analyzing image');
  } catch (e) {
    console.error('AI analysis failed:', e);
    showToast('AI error: ' + e.message, 'error');
    return null;
  }
}

async function generateTagsWithAI(description) {
  if (!aiAvailable) return [];
  try {
    const result = await submitAIJob('tags', { description }, 'Generating tags');
    return result.tags || [];
  } catch (e) {
    console.error('Tag generation failed:', e);
    return [];
  }
}

async function dressWithAI(bodyImage, clothingItems, customPrompt = '') {
  if (!aiAvailable) return null;
  try {
    return await submitAIJob('dress', {
      body_image: bodyImage, clothing_items: clothingItems, prompt: customPrompt
    }, 'Positioning clothing');
  } catch (e) {
    console.error('AI dress failed:', e);
    showToast('AI dress failed: ' + e.message, 'error');
    return null;
  }
}

async function removeBackgroundWithAI(imageDataUrl) {
  try {
    const result = await submitAIJob('remove-bg', { image: imageDataUrl }, 'Removing background');
    return result.image || null;
  } catch (e) {
    console.error('Background removal failed:', e);
    showToast('Background removal failed: ' + e.message, 'error');
    return null;
  }
}

async function suggestOutfitWithAI(occasion, weather, items) {
  if (!aiAvailable) return null;
  try {
    return await submitAIJob('suggest-outfit', { occasion, weather, items }, 'Suggesting outfit');
  } catch (e) {
    console.error('Outfit suggestion failed:', e);
    showToast('Outfit suggestion failed: ' + e.message, 'error');
    return null;
  }
}

async function analyzeColorMatch(items) {
  if (!aiAvailable) return null;
  try {
    return await submitAIJob('color-match', { items }, 'Analyzing colors');
  } catch (e) {
    console.error('Color match failed:', e);
    return null;
  }
}

async function findStyleMatches(baseItem, items, maxItems = 5) {
  if (!aiAvailable) return null;
  try {
    return await submitAIJob('style-match', { base_item: baseItem, items, max_items: maxItems }, 'Finding matches');
  } catch (e) {
    console.error('Style match failed:', e);
    return null;
  }
}

async function rateOutfitWithAI(items) {
  if (!aiAvailable) return null;
  try {
    return await submitAIJob('rate-outfit', { items }, 'Rating outfit');
  } catch (e) {
    console.error('Outfit rating failed:', e);
    return null;
  }
}

async function analyzeWardrobeGaps(items) {
  if (!aiAvailable) return null;
  try {
    return await submitAIJob('wardrobe-gaps', { items }, 'Analyzing wardrobe');
  } catch (e) {
    console.error('Wardrobe gaps failed:', e);
    return null;
  }
}

// ============================================================================
// DATA STORE (Mock localStorage persistence)
// ============================================================================

const STORAGE_KEY = 'gothreads_mockup_data';


const defaultData = {
  items: [],
  outfits: [],
  wearHistory: [],
  calendarEvents: {},
  sharedOutfits: []
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
let _editingOutfitId = null; // Set when editing an existing outfit (DB id)

// Sync items from API into appData so all item lookups work.
// The API (PostgreSQL) is the source of truth for items.
// localStorage still stores outfits, wear history, calendar, etc.
async function syncItemsFromAPI() {
  try {
    const response = await fetch('http://localhost:8556/api/items');
    if (!response.ok) return;
    const apiItems = await response.json();
    // Normalize: API returns image_url, local code expects image
    appData.items = apiItems.map(item => ({
      ...item,
      image: item.image_url || item.image
    }));
    saveData(appData);
  } catch (e) {
    console.log('Could not sync items from API:', e.message);
  }
}

// Sync outfits from API into appData
async function syncOutfitsFromAPI() {
  try {
    const response = await fetch('http://localhost:8556/api/outfits');
    if (!response.ok) return;
    const apiOutfits = await response.json();
    // Normalize: API returns item_ids/body_image_url, local code expects itemIds/bodyBackground
    appData.outfits = apiOutfits.map(o => ({
      id: o.id,
      name: o.name,
      notes: o.notes || '',
      itemIds: o.item_ids || [],
      positions: (o.positions || []).map(p => ({ id: p.id, x: p.x, y: p.y, z: p.z })),
      bodyBackground: o.body_image_url ? `url(${o.body_image_url})` : null,
      wearCount: o.wear_count || 0,
      rating: o.rating || 0,
      createdAt: o.created_at,
    }));
    saveData(appData);
  } catch (e) {
    console.log('Could not sync outfits from API:', e.message);
  }
}

// Sync calendar events from API into appData
async function syncCalendarFromAPI() {
  try {
    const response = await fetch('http://localhost:8556/api/calendar');
    if (!response.ok) return;
    const apiEvents = await response.json();
    // Convert array to date-keyed object
    const calendarEvents = {};
    apiEvents.forEach(e => {
      calendarEvents[e.event_date] = {
        id: e.id,
        outfitId: e.outfit_id || null,
        eventName: e.event_name || '',
        weather: e.weather || '',
        location: e.location || '',
        date: e.event_date,
      };
    });
    appData.calendarEvents = calendarEvents;
    saveData(appData);
  } catch (e) {
    console.log('Could not sync calendar from API:', e.message);
  }
}

// One-time migration of localStorage outfits and calendar events to API
async function migrateLocalStorageToAPI() {
  if (localStorage.getItem('gothreads_migrated_to_api')) return;

  // Migrate outfits
  const localOutfits = appData.outfits || [];
  for (const outfit of localOutfits) {
    // Skip if already has a small numeric ID (likely from DB)
    if (outfit.id < 1000000) continue;
    try {
      const bodyUrl = outfit.bodyBackground ? outfit.bodyBackground.replace(/^url\(["']?/, '').replace(/["']?\)$/, '') : '';
      const resp = await fetch('http://localhost:8556/api/outfits', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          name: outfit.name,
          notes: outfit.notes || '',
          body_image_url: bodyUrl,
          positions: (outfit.positions || outfit.itemIds?.map(id => ({ id, x: 100, y: 100, z: 0 })) || []),
        })
      });
      if (resp.ok) {
        console.log(`Migrated outfit: ${outfit.name}`);
      }
    } catch (e) {
      console.log(`Failed to migrate outfit ${outfit.name}:`, e.message);
    }
  }

  // Migrate calendar events
  const localEvents = appData.calendarEvents || {};
  for (const [dateStr, event] of Object.entries(localEvents)) {
    try {
      await fetch('http://localhost:8556/api/calendar', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          date: dateStr,
          event_name: event.eventName || '',
          weather: event.weather || '',
          location: event.location || '',
          outfit_id: event.outfitId || null,
        })
      });
    } catch (e) {
      console.log(`Failed to migrate calendar event ${dateStr}:`, e.message);
    }
  }

  localStorage.setItem('gothreads_migrated_to_api', 'true');
  console.log('localStorage migration to API complete');
}

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
  const validTypes = ['image/jpeg', 'image/png', 'image/webp' ];
  if (!validTypes.includes(file.type)) {
    showToast('Please upload a JPEG, PNG, or WebP image', 'error');
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

        // Upload image to S3
        updateProcessingStatus('Uploading image to storage...');
        let imageURL = imageData;
        try {
          const uploadResponse = await fetch('http://localhost:8556/api/upload', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ image: imageData })
          });

          if (uploadResponse.ok) {
            const uploadResult = await uploadResponse.json();
            imageURL = uploadResult.url;
            console.log('Image uploaded to S3:', imageURL);
          } else {
            console.warn('S3 upload failed, falling back to base64');
          }
        } catch (uploadError) {
          console.warn('S3 upload error, falling back to base64:', uploadError);
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

        // Check if AI analysis has real data (not defaults)
        const hasValidAIData = aiResult &&
                               aiResult.category &&
                               aiResult.category !== 'Unknown' &&
                               aiResult.description &&
                               aiResult.description !== 'Unable to analyze';

        // Save to database via API
        const newItem = {
          name: hasValidAIData ? (aiResult.name || aiResult.description.substring(0, 50)) : 'Unnamed Item',
          description: hasValidAIData ? (aiResult.description || '') : '',
          category: hasValidAIData ? aiResult.category : 'Uncategorized',
          price: 0,
          color: hasValidAIData ? (aiResult.color || '') : '',
          brand: hasValidAIData ? (aiResult.brand || '') : '',
          image_url: imageURL,
          tags: hasValidAIData ? (aiResult.tags || []) : [],
          ai_analysis: hasValidAIData ? (aiResult.raw_response || null) : null,
        };

        try {
          const response = await fetch('http://localhost:8556/api/items', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(newItem)
          });

          if (!response.ok) {
            throw new Error('Failed to save item to database');
          }

          const savedItem = await response.json();

          if (shouldRedirect) {
            closeModal();

            if (hasValidAIData) {
              const brandText = aiResult.brand ? ` by ${aiResult.brand}` : '';
              showToast(`AI detected: ${aiResult.color} ${aiResult.category}${brandText}`, 'success');
            } else {
              showToast('⚠️ AI could not analyze item. Please enter details manually.', 'info');
            }

            // Redirect to edit page
            setTimeout(() => {
              window.location.href = `item-detail.html?id=${savedItem.id}`;
            }, 1000);
          }

          resolve(true);
        } catch (apiError) {
          console.error('API save error:', apiError);
          showToast('Failed to save item: ' + apiError.message, 'error');
          resolve(false);
        }
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

async function renderWardrobe(containerId, filterCategory = 'all') {
  const container = document.getElementById(containerId);
  if (!container) return;

  // Check if we should show recent uploads
  const urlParams = new URLSearchParams(window.location.search);
  const showRecent = urlParams.get('recent') === 'true';

  // Fetch items from API
  let items = [];
  try {
    const response = await fetch('http://localhost:8556/api/items');
    if (response.ok) {
      items = await response.json();
    }
  } catch (error) {
    console.error('Failed to fetch items:', error);
  }

  if (filterCategory !== 'all') {
    items = items.filter(item => item.category === filterCategory);
  }

  // Sort by upload date if showing recent
  if (showRecent) {
    items = items.sort((a, b) => new Date(b.uploaded_at || 0) - new Date(a.uploaded_at || 0));
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

    const wearCount = item.wear_count || 0;
    const price = item.price || 0;
    const costPerWear = wearCount > 0 ? (price / wearCount).toFixed(2) : price.toFixed(2);

    // Highlight AI-detected info or show retry button
    const aiInfo = item.ai_analysis ? `
      <div style="font-size: 0.75rem; color: var(--primary); margin-top: 0.25rem;">
        ✨ ${item.color}${item.brand ? ` • ${item.brand}` : ''}
      </div>
    ` : `
      <div style="font-size: 0.75rem; color: var(--text-neutral); margin-top: 0.25rem;">
        ⚠️ No AI metadata
      </div>
    `;

    const needsAnalysis = !item.ai_analysis || item.ai_analysis === '' || item.name === 'Unnamed Item' || item.category === 'Uncategorized';
    const retryButton = needsAnalysis ? `
      <div id="retry-container-${item.id}">
        <button class="btn btn-secondary" style="width: 100%; margin-top: 0.5rem; padding: 0.5rem; font-size: 0.75rem;"
                onclick="event.stopPropagation(); retryAIFromWardrobe(${item.id});">
          🔄 Analyze with AI
        </button>
      </div>
    ` : '';

    card.innerHTML = `
      <img src="${item.image_url}" alt="${item.name}" class="item-image" loading="lazy">
      <div class="item-info">
        <div class="item-name">${item.name}</div>
        <div class="item-meta">${item.category} • $${price.toFixed(2)}</div>
        ${aiInfo}
        <div class="item-stats">
          <span title="Times worn">👔 ${wearCount}</span>
          <span title="Cost per wear">💰 $${costPerWear}</span>
        </div>
        ${retryButton}
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
  const validTypes = ['image/jpeg', 'image/png', 'image/webp' ];
  if (!validTypes.includes(file.type)) {
    showToast('Please upload a JPEG, PNG, or WebP image', 'error');
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
    <img src="${item.image}" alt="${item.name}" draggable="false" style="width: 100%; height: 100%; object-fit: contain;">
    <div class="canvas-item-controls">
      <button onclick="resizeCanvasItem(this.closest('.canvas-item'), 'bigger')" title="Make bigger">+</button>
      <button onclick="resizeCanvasItem(this.closest('.canvas-item'), 'smaller')" title="Make smaller">−</button>
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

function resizeCanvasItem(element, direction) {
  const currentWidth = element.offsetWidth;
  const currentHeight = element.offsetHeight;

  // Resize by 20px increments
  const change = direction === 'bigger' ? 20 : -20;
  const newWidth = Math.max(50, currentWidth + change); // Minimum 50px
  const newHeight = Math.max(50, currentHeight + change);

  element.style.width = `${newWidth}px`;
  element.style.height = `${newHeight}px`;

  // Update in currentOutfitItems
  const itemId = parseInt(element.dataset.itemId);
  const item = currentOutfitItems.find(i => i.id === itemId);
  if (item) {
    item.width = newWidth;
    item.height = newHeight;
  }
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

async function getWeatherForLocation(lat = 51.5074, lon = -0.1278) {
  // Default: London coordinates
  try {
    const response = await fetch(`https://api.open-meteo.com/v1/forecast?latitude=${lat}&longitude=${lon}&current=temperature_2m,weather_code&daily=temperature_2m_max,temperature_2m_min&timezone=auto`);
    const data = await response.json();

    const temp = data.current.temperature_2m;
    const weatherCode = data.current.weather_code;

    // Weather code to description mapping
    const weatherDesc = weatherCode <= 3 ? 'clear' :
                       weatherCode <= 67 ? 'rainy' :
                       weatherCode <= 77 ? 'snowy' :
                       weatherCode <= 99 ? 'stormy' : 'mild';

    // Temperature to weather condition
    const condition = temp < 5 ? 'cold' : temp < 15 ? 'mild' : temp < 25 ? 'warm' : 'hot';

    // Determine season based on month
    const month = new Date().getMonth();
    const season = month >= 2 && month <= 4 ? 'spring' :
                  month >= 5 && month <= 7 ? 'summer' :
                  month >= 8 && month <= 10 ? 'autumn' : 'winter';

    return {
      temperature: temp,
      condition,
      weather: weatherDesc,
      season
    };
  } catch (error) {
    console.error('Weather fetch error:', error);
    // Fallback
    return {
      temperature: 15,
      condition: 'mild',
      weather: 'mild',
      season: 'spring'
    };
  }
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

  // If no items on canvas, generate outfit suggestion first
  if (currentOutfitItems.length === 0) {
    await generateOutfitSuggestion();
    return;
  }

  // Position existing items on canvas
  await positionItemsWithAI();
}

async function handleTryOn() {
  if (!aiAvailable) {
    showToast('AI not available. Start the API server.', 'error');
    return;
  }

  const canvas = document.querySelector('.outfit-canvas');
  const bodyBg = canvas?.style.backgroundImage;

  if (!bodyBg || bodyBg === 'none') {
    showToast('Please add a body/mannequin image first!', 'error');
    return;
  }

  if (currentOutfitItems.length === 0) {
    showToast('Add at least one clothing item to try on!', 'error');
    return;
  }

  // Use the first selected item, or first item on canvas
  const selectedEl = document.querySelector('.canvas-item.selected');
  const targetItemId = selectedEl ? parseInt(selectedEl.dataset.itemId) : currentOutfitItems[0].id;
  const targetItem = currentOutfitItems.find(i => i.id === targetItemId);

  if (!targetItem) {
    showToast('Could not find selected item', 'error');
    return;
  }

  // Determine cloth_type from category
  const normCat = normalizeCategory(targetItem.category);
  let clothType = 'upper';
  if (normCat === 'Bottoms') clothType = 'lower';
  if (normCat === 'Outerwear') clothType = 'overall';

  try {
    // Extract body image base64 from background
    const bgUrl = bodyBg.replace(/^url\(["']?/, '').replace(/["']?\)$/, '');
    let personImage = bgUrl;

    // Convert SVG to PNG via canvas (VTON models need raster images)
    if (bgUrl.includes('data:image/svg')) {
      personImage = await new Promise((resolve, reject) => {
        const img = new window.Image();
        img.onload = () => {
          const c = document.createElement('canvas');
          c.width = 768;
          c.height = 1024;
          const ctx = c.getContext('2d');
          ctx.fillStyle = '#F5F0E8';
          ctx.fillRect(0, 0, c.width, c.height);
          const scale = Math.min(c.width / img.width, c.height / img.height);
          const x = (c.width - img.width * scale) / 2;
          const y = (c.height - img.height * scale) / 2;
          ctx.drawImage(img, x, y, img.width * scale, img.height * scale);
          resolve(c.toDataURL('image/png'));
        };
        img.onerror = reject;
        img.src = bgUrl;
      });
    } else if (bgUrl.startsWith('blob:') || bgUrl.startsWith('http')) {
      const resp = await fetch(bgUrl);
      const blob = await resp.blob();
      personImage = await new Promise(resolve => {
        const reader = new FileReader();
        reader.onload = () => resolve(reader.result);
        reader.readAsDataURL(blob);
      });
    }

    // Get garment image (canvas items use .image, API items use .image_url)
    let garmentImage = targetItem.image || targetItem.image_url;
    if (garmentImage && !garmentImage.startsWith('data:')) {
      const resp = await fetch(garmentImage);
      const blob = await resp.blob();
      garmentImage = await new Promise(resolve => {
        const reader = new FileReader();
        reader.onload = () => resolve(reader.result);
        reader.readAsDataURL(blob);
      });
    }

    // Submit via generic AI job queue (SSE progress in notification panel)
    const result = await submitAIJob('tryon', {
      person_image: personImage,
      garment_image: garmentImage,
      cloth_type: clothType
    }, `Try-on: ${targetItem.name || 'item'}`);

    // Show result as the canvas background
    canvas.style.backgroundImage = `url(${result.image})`;
    canvas.style.backgroundSize = 'contain';
    canvas.style.backgroundPosition = 'center';
    canvas.style.backgroundRepeat = 'no-repeat';

    showToast('Virtual try-on complete!', 'success');

  } catch (err) {
    console.error('Try-on error:', err);
    if (err.message !== 'Cancelled') {
      showToast(`Try-on failed: ${err.message}`, 'error');
    }
  }
}

async function generateOutfitSuggestion() {
  try {
    // Fetch weather for London
    const weather = await getWeatherForLocation(51.5074, -0.1278);

    // Get current date/time info
    const now = new Date();
    const dateStr = now.toLocaleDateString('en-GB', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' });
    const timeStr = now.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' });

    // Fetch all items from wardrobe
    const response = await fetch('http://localhost:8556/api/items');
    if (!response.ok) {
      throw new Error('Failed to fetch wardrobe items');
    }
    const allItems = await response.json();

    if (!allItems || allItems.length === 0) {
      showToast('No items in wardrobe. Upload some clothing first!', 'error');
      return;
    }

    // Prepare items for AI with full context
    const itemsForAI = allItems.map(item => ({
      id: item.id,
      name: item.name,
      category: item.category,
      color: item.color,
      brand: item.brand || '',
      tags: item.tags || [],
      description: item.ai_analysis || '',
      wear_count: item.wear_count || 0
    }));

    // Get selected occasion or randomize
    const occasionSelect = document.getElementById('occasion-select');
    let occasion = occasionSelect ? occasionSelect.value : 'casual';

    // If "Surprise Me" is selected, randomize the occasion
    if (occasion === 'random') {
      const occasions = ['casual', 'work', 'formal', 'date', 'party', 'outdoor'];
      occasion = occasions[Math.floor(Math.random() * occasions.length)];
      console.log('🎲 Randomized occasion:', occasion);
    }

    const weatherDesc = `${weather.condition}, ${weather.weather}, ${weather.season}`;
    const suggestion = await suggestOutfitWithAI(occasion, weatherDesc, itemsForAI);

    if (!suggestion || !suggestion.selected_items || suggestion.selected_items.length === 0) {
      showToast('AI could not suggest an outfit', 'error');
      return;
    }

    // Show weather info card on the page
    displayWeatherInfo(weather, dateStr, timeStr, suggestion, occasion);

    // Add suggested items to canvas
    for (const itemId of suggestion.selected_items) {
      const item = allItems.find(i => i.id === itemId);
      if (item) {
        // Convert to old format for compatibility
        const canvasItem = {
          id: item.id,
          name: item.name,
          category: item.category,
          image: item.image_url,
          color: item.color,
          brand: item.brand
        };
        addToCanvas(canvasItem);
      }
    }

    showToast(`✨ Outfit suggested for ${weather.condition} weather! Check the info box below.`, 'success');

    // Store context for chat refinements
    window.currentOutfitContext = {
      weather,
      occasion,
      allItems,
      selectedItems: suggestion.selected_items,
      reasoning: suggestion.reasoning
    };

    // Show refinement options
    showOutfitRefinementOptions();

    // Wait a bit then position the items
    setTimeout(async () => {
      await positionItemsWithAI();
    }, 1500);

  } catch (error) {
    console.error('Outfit generation error:', error);
    showToast('Error generating outfit: ' + error.message, 'error');
  }
}

function displayWeatherInfo(weather, dateStr, timeStr, suggestion, occasion) {
  // Remove existing info box if present
  const existingBox = document.getElementById('outfit-context-info');
  if (existingBox) {
    existingBox.remove();
  }

  // Create info box
  const infoBox = document.createElement('div');
  infoBox.id = 'outfit-context-info';
  infoBox.style.cssText = `
    margin-top: 1.5rem;
    padding: 1.5rem;
    background: linear-gradient(135deg, #E8F5E9 0%, #C8E6C9 100%);
    border-radius: 0.75rem;
    border: 2px solid #4CAF50;
  `;

  const weatherEmoji = weather.condition === 'cold' ? '❄️' :
                      weather.condition === 'hot' ? '☀️' :
                      weather.condition === 'warm' ? '🌤️' : '🌥️';

  const seasonEmoji = weather.season === 'winter' ? '❄️' :
                     weather.season === 'spring' ? '🌸' :
                     weather.season === 'summer' ? '☀️' : '🍂';

  const occasionEmoji = occasion === 'work' ? '💼' :
                       occasion === 'formal' ? '🎩' :
                       occasion === 'date' ? '💕' :
                       occasion === 'gym' ? '🏋️' :
                       occasion === 'beach' ? '🏖️' :
                       occasion === 'party' ? '🎉' :
                       occasion === 'outdoor' ? '🏕️' : '👕';

  infoBox.innerHTML = `
    <h3 style="margin: 0 0 1rem 0; color: #2E7D32; display: flex; align-items: center; gap: 0.5rem;">
      <span>🤖</span> AI Outfit Context
    </h3>

    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 1rem; margin-bottom: 1rem;">
      <div style="background: white; padding: 1rem; border-radius: 0.5rem;">
        <div style="font-size: 0.75rem; color: #666; margin-bottom: 0.25rem;">${occasionEmoji} Occasion</div>
        <div style="font-weight: 600; text-transform: capitalize;">${occasion}</div>
      </div>

      <div style="background: white; padding: 1rem; border-radius: 0.5rem;">
        <div style="font-size: 0.75rem; color: #666; margin-bottom: 0.25rem;">📍 Location</div>
        <div style="font-weight: 600;">London, UK</div>
      </div>

      <div style="background: white; padding: 1rem; border-radius: 0.5rem;">
        <div style="font-size: 0.75rem; color: #666; margin-bottom: 0.25rem;">📅 Date & Time</div>
        <div style="font-weight: 600; font-size: 0.9rem;">${dateStr}</div>
        <div style="font-size: 0.875rem; color: #666;">${timeStr}</div>
      </div>

      <div style="background: white; padding: 1rem; border-radius: 0.5rem;">
        <div style="font-size: 0.75rem; color: #666; margin-bottom: 0.25rem;">🌡️ Temperature</div>
        <div style="font-weight: 600; font-size: 1.5rem;">${weather.temperature}°C</div>
        <div style="font-size: 0.875rem; color: #666;">${weatherEmoji} ${weather.condition}</div>
      </div>

      <div style="background: white; padding: 1rem; border-radius: 0.5rem;">
        <div style="font-size: 0.75rem; color: #666; margin-bottom: 0.25rem;">🌦️ Conditions</div>
        <div style="font-weight: 600;">${weather.weather}</div>
        <div style="font-size: 0.875rem; color: #666;">${seasonEmoji} ${weather.season}</div>
      </div>
    </div>

    <div style="background: white; padding: 1rem; border-radius: 0.5rem; border-left: 4px solid #4CAF50;">
      <div style="font-size: 0.875rem; color: #666; margin-bottom: 0.5rem;">💬 AI Reasoning</div>
      <p style="margin: 0; line-height: 1.6;">${suggestion.reasoning || 'AI selected items based on current weather conditions and season.'}</p>
      ${suggestion.tips && suggestion.tips.length > 0 ? `
        <div style="margin-top: 0.75rem;">
          <div style="font-size: 0.75rem; color: #666; margin-bottom: 0.25rem;">💡 Style Tips</div>
          <ul style="margin: 0; padding-left: 1.25rem; line-height: 1.8;">
            ${suggestion.tips.map(tip => `<li style="font-size: 0.875rem;">${tip}</li>`).join('')}
          </ul>
        </div>
      ` : ''}
    </div>
  `;

  // Insert after the outfit canvas section
  const canvasContainer = document.querySelector('.outfit-canvas').closest('.card');
  canvasContainer.parentNode.insertBefore(infoBox, canvasContainer.nextSibling);

  // Scroll to show the info
  infoBox.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
}

function showOutfitRefinementOptions() {
  // Remove existing refinement box if present
  const existingBox = document.getElementById('outfit-refinement-box');
  if (existingBox) {
    existingBox.remove();
  }

  const refinementBox = document.createElement('div');
  refinementBox.id = 'outfit-refinement-box';
  refinementBox.style.cssText = `
    margin-top: 1rem;
    padding: 1.5rem;
    background: linear-gradient(135deg, #FFF5F4 0%, #FFE8E6 100%);
    border-radius: 0.75rem;
    border: 2px solid var(--primary);
  `;

  refinementBox.innerHTML = `
    <h3 style="margin: 0 0 1rem 0; color: var(--primary); display: flex; align-items: center; gap: 0.5rem;">
      <span>✨</span> Want to adjust this outfit?
    </h3>

    <div style="display: flex; gap: 0.75rem; flex-wrap: wrap; margin-bottom: 1rem;">
      <button class="btn btn-secondary" onclick="refineOutfit('different-color')" style="font-size: 0.875rem;">
        🎨 Different Color
      </button>
      <button class="btn btn-secondary" onclick="refineOutfit('more-formal')" style="font-size: 0.875rem;">
        👔 More Formal
      </button>
      <button class="btn btn-secondary" onclick="refineOutfit('more-casual')" style="font-size: 0.875rem;">
        👕 More Casual
      </button>
      <button class="btn btn-secondary" onclick="refineOutfit('warmer')" style="font-size: 0.875rem;">
        🧥 Add Layer
      </button>
      <button class="btn btn-secondary" onclick="refineOutfit('lighter')" style="font-size: 0.875rem;">
        ☀️ Remove Layer
      </button>
    </div>

    <div style="display: flex; gap: 0.5rem; align-items: center;">
      <input type="text" id="custom-refinement-input" placeholder="Or describe what you'd like to change..."
             style="flex: 1; padding: 0.75rem; border: 2px solid var(--border); border-radius: 0.5rem; font-size: 0.875rem;">
      <button class="btn btn-primary" onclick="refineOutfitCustom()" style="white-space: nowrap;">
        🤖 Ask AI
      </button>
    </div>
  `;

  // Insert after context info box
  const contextBox = document.getElementById('outfit-context-info');
  if (contextBox) {
    contextBox.parentNode.insertBefore(refinementBox, contextBox.nextSibling);
  }
}

async function refineOutfit(refinementType) {
  if (!window.currentOutfitContext) {
    showToast('No outfit to refine. Generate an outfit first!', 'error');
    return;
  }

  const refinementMessages = {
    'different-color': 'Use different colors',
    'more-formal': 'Make it more formal',
    'more-casual': 'Make it more casual',
    'warmer': 'Add a warmer layer',
    'lighter': 'Remove a layer, make it lighter'
  };

  const message = refinementMessages[refinementType];
  await refineOutfitWithAI(message);
}

async function refineOutfitCustom() {
  const input = document.getElementById('custom-refinement-input');
  const message = input.value.trim();

  if (!message) {
    showToast('Please describe what you want to change', 'error');
    return;
  }

  await refineOutfitWithAI(message);
  input.value = '';
}

async function refineOutfitWithAI(refinementMessage) {
  const ctx = window.currentOutfitContext;

  try {
    const suggestion = await suggestOutfitWithAI(
      ctx.occasion,
      `${ctx.weather.condition}, ${ctx.weather.weather}, ${ctx.weather.season}. User wants: ${refinementMessage}`,
      ctx.allItems.map(item => ({
        id: item.id,
        name: item.name,
        category: item.category,
        color: item.color,
        brand: item.brand || '',
        tags: item.tags || []
      }))
    );

    if (!suggestion || !suggestion.selected_items || suggestion.selected_items.length === 0) {
      showToast('AI could not refine the outfit', 'error');
      return;
    }

    // Clear current items from canvas
    currentOutfitItems.forEach(item => {
      const element = document.querySelector(`.canvas-item[data-item-id="${item.id}"]`);
      if (element) element.remove();
    });
    currentOutfitItems = [];

    // Add new suggested items
    for (const itemId of suggestion.selected_items) {
      const item = ctx.allItems.find(i => i.id === itemId);
      if (item) {
        const canvasItem = {
          id: item.id,
          name: item.name,
          category: item.category,
          image: item.image_url,
          color: item.color,
          brand: item.brand
        };
        addToCanvas(canvasItem);
      }
    }

    // Update context
    ctx.selectedItems = suggestion.selected_items;
    ctx.reasoning = suggestion.reasoning;

    // Update info display
    displayWeatherInfo(ctx.weather, new Date().toLocaleDateString('en-GB', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' }),
                      new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' }),
                      suggestion, ctx.occasion);

    showToast('✨ Outfit refined! Now positioning...', 'success');

    // Position items
    setTimeout(async () => {
      await positionItemsWithAI();
    }, 1500);

  } catch (error) {
    closeModal();
    console.error('Refinement error:', error);
    showToast('Error refining outfit: ' + error.message, 'error');
  }
}

// Map variant category names to canonical ones for canvas positioning
// Supports layering: Tops (base) -> Midlayer (sweater/shacket) -> Outerwear (coat/jacket)
function normalizeCategory(cat) {
  if (!cat) return 'Accessories';
  const c = cat.toLowerCase();
  if (['tops', 'shirts', 't-shirts', 'blouses'].includes(c)) return 'Tops';
  if (['sweaters', 'knitwear', 'cardigans', 'hoodies', 'shackets', 'fleece'].includes(c)) return 'Midlayer';
  if (['bottoms', 'pants', 'jeans', 'trousers', 'shorts', 'skirts'].includes(c)) return 'Bottoms';
  if (['shoes', 'boots', 'sneakers', 'sandals', 'footwear'].includes(c)) return 'Shoes';
  if (['outerwear', 'jackets', 'coats', 'blazers', 'parkas', 'windbreakers'].includes(c)) return 'Outerwear';
  if (['accessories', 'hats', 'bags', 'belts', 'scarves', 'watches', 'jewelry'].includes(c)) return 'Accessories';
  return 'Accessories';
}

const categorySizeMap = {
  'Tops':       { w: 140, h: 140 },
  'Midlayer':   { w: 145, h: 145 },
  'Bottoms':    { w: 130, h: 150 },
  'Shoes':      { w: 110, h: 100 },
  'Outerwear':  { w: 150, h: 160 },
  'Accessories': { w: 80, h: 80 },
};

function positionItemsByCategory() {
  const canvas = document.querySelector('.outfit-canvas');
  if (!canvas) return;

  const canvasRect = canvas.getBoundingClientRect();
  const cw = canvasRect.width;
  const ch = canvasRect.height;

  const categoryPositions = {
    'Tops':        { x: 0.38, y: 0.05, w: 140, h: 140, z: 2 },
    'Midlayer':    { x: 0.35, y: 0.04, w: 145, h: 145, z: 3 },
    'Bottoms':     { x: 0.38, y: 0.38, w: 130, h: 150, z: 1 },
    'Shoes':       { x: 0.38, y: 0.72, w: 110, h: 100, z: 1 },
    'Outerwear':   { x: 0.05, y: 0.03, w: 150, h: 160, z: 4 },
    'Accessories': { x: 0.75, y: 0.10, w: 80,  h: 80,  z: 5 },
  };

  const categoryCount = {};

  currentOutfitItems.forEach(item => {
    const el = document.querySelector(`.canvas-item[data-item-id="${item.id}"]`);
    if (!el) return;

    const cat = normalizeCategory(item.category);
    const pos = categoryPositions[cat];
    const size = { w: pos.w, h: pos.h };

    const count = categoryCount[cat] || 0;
    categoryCount[cat] = count + 1;
    const offsetX = count * 30;
    const offsetY = count * 15;

    const x = pos.x * cw + offsetX;
    const y = pos.y * ch + offsetY;

    el.style.left = `${x}px`;
    el.style.top = `${y}px`;
    el.style.width = `${size.w}px`;
    el.style.height = `${size.h}px`;
    el.style.zIndex = pos.z;

    item.x = x;
    item.y = y;
    item.z = pos.z;
  });
}

async function positionItemsWithAI() {
  const canvas = document.querySelector('.outfit-canvas');
  if (!canvas) return;

  const bodyBg = canvas.style.backgroundImage;
  const match = bodyBg?.match(/url\("?(.+?)"?\)/);
  if (!match) {
    // No body image, fall back to category-based
    positionItemsByCategory();
    return;
  }
  const bodyDataUrl = match[1];

  const clothingData = currentOutfitItems.map(item => ({
    id: item.id,
    name: item.name,
    category: item.category,
    image: item.image
  }));

  try {
    const result = await dressWithAI(bodyDataUrl, clothingData);

    if (result && result.positions && result.positions.length > 0) {
      console.log('AI Dress result:', result);

      const canvasRect = canvas.getBoundingClientRect();

      result.positions.forEach(pos => {
        const canvasItem = document.querySelector(`.canvas-item[data-item-id="${pos.id}"]`);
        if (!canvasItem) return;

        const item = currentOutfitItems.find(i => i.id === pos.id);
        const cat = normalizeCategory(item?.category);
        const size = categorySizeMap[cat] || categorySizeMap['Accessories'];

        // Convert percentage to pixels, offset by half the item size to center it
        const x = (pos.x / 100) * canvasRect.width - size.w / 2;
        const y = (pos.y / 100) * canvasRect.height - size.h / 2;

        canvasItem.style.left = `${Math.max(0, x)}px`;
        canvasItem.style.top = `${Math.max(0, y)}px`;
        canvasItem.style.width = `${size.w}px`;
        canvasItem.style.height = `${size.h}px`;
        canvasItem.style.zIndex = pos.z;

        if (item) {
          item.x = x;
          item.y = y;
          item.z = pos.z;
        }
      });

      showToast(result.explanation || 'AI positioned clothing on body!', 'success');
    } else {
      // Fallback to category-based
      console.log('AI positioning returned no results, using category-based fallback');
      positionItemsByCategory();
      showToast('Positioned items by category', 'info');
    }
  } catch (e) {
    console.error('AI dress failed, falling back to category-based:', e);
    positionItemsByCategory();
    showToast('AI positioning unavailable, used category layout', 'info');
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

  // Prepare items for rating (include description/ai_analysis for accurate feedback)
  const items = currentOutfitItems.map(item => ({
    id: item.id,
    name: item.name,
    category: item.category,
    color: item.color,
    tags: item.tags || [],
    description: item.description || item.ai_analysis || ''
  }));

  try {
    const result = await rateOutfitWithAI(items);

    if (result) {
      const ratingStars = '⭐'.repeat(Math.round(result.rating / 2));
      const strengthsList = result.strengths && result.strengths.length > 0 ?
        `<ul>${result.strengths.map(s => `<li>${s}</li>`).join('')}</ul>` : '';
      const improvementsList = result.improvements && result.improvements.length > 0 ?
        `<ul>${result.improvements.map(i => `<li>${i}</li>`).join('')}</ul>` : '';

      const colorScoreColor = result.color_score >= 80 ? '#27AE60' :
                              result.color_score >= 60 ? '#F39C12' : '#E74C3C';

      showModal('⭐ Outfit Rating', `
        <div style="text-align: center; margin: 1.5rem 0;">
          <div style="font-size: 3rem;">${ratingStars}</div>
          <div style="font-size: 2rem; font-weight: bold; color: var(--primary);">${result.rating}/10</div>
        </div>
        <div style="padding: 1rem; background: var(--warm-bg); border-radius: 0.5rem; margin-bottom: 1rem;">
          <p style="margin: 0; color: var(--text-neutral);">${result.feedback}</p>
        </div>
        ${result.color_score ? `
          <div style="padding: 1rem; background: var(--warm-bg); border-radius: 0.5rem; margin-bottom: 1rem;">
            <div style="display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.5rem;">
              <strong>🎨 Color Harmony:</strong>
              <span style="font-size: 1.25rem; font-weight: bold; color: ${colorScoreColor};">${result.color_score}/100</span>
            </div>
            ${result.color_analysis ? `<p style="margin: 0; color: var(--text-neutral); font-size: 0.875rem;">${result.color_analysis}</p>` : ''}
          </div>
        ` : ''}
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
      showToast('AI could not rate outfit', 'error');
    }
  } catch (e) {
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

async function confirmSaveOutfit() {
  const nameInput = document.getElementById('outfit-name-input');
  const notesInput = document.getElementById('outfit-notes-input');

  const name = nameInput?.value.trim();
  if (!name) {
    showToast('Please enter a name for your outfit', 'error');
    return;
  }

  // Get body background if set
  const canvas = document.querySelector('.outfit-canvas');
  const bgImage = canvas?.style.backgroundImage;
  const bodyImageUrl = bgImage && bgImage !== 'none' ? bgImage.replace(/^url\(["']?/, '').replace(/["']?\)$/, '') : '';

  const positions = currentOutfitItems.map(i => ({
    id: i.id,
    x: Math.round(i.x || 0),
    y: Math.round(i.y || 0),
    z: i.z || 0,
    scale: 1.0
  }));

  const payload = {
    name: name,
    notes: notesInput?.value.trim() || '',
    body_image_url: bodyImageUrl,
    positions: positions,
  };

  try {
    let resp;
    if (_editingOutfitId) {
      // Update existing outfit
      resp = await fetch(`http://localhost:8556/api/outfits/${_editingOutfitId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
    } else {
      // Create new outfit
      resp = await fetch('http://localhost:8556/api/outfits', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
      });
    }

    if (!resp.ok) {
      throw new Error('API request failed');
    }

    const result = await resp.json();

    // Update local state
    const localOutfit = {
      id: _editingOutfitId || result.id,
      name: name,
      notes: notesInput?.value.trim() || '',
      itemIds: currentOutfitItems.map(i => i.id),
      positions: currentOutfitItems.map(i => ({ id: i.id, x: i.x, y: i.y, z: i.z })),
      bodyBackground: bgImage !== 'none' ? bgImage : null,
      wearCount: 0,
      rating: 0,
      createdAt: new Date().toISOString(),
    };

    if (_editingOutfitId) {
      const idx = appData.outfits.findIndex(o => o.id == _editingOutfitId);
      if (idx >= 0) {
        localOutfit.wearCount = appData.outfits[idx].wearCount || 0;
        localOutfit.rating = appData.outfits[idx].rating || 0;
        appData.outfits[idx] = localOutfit;
      }
    } else {
      localOutfit.id = result.id;
      appData.outfits.push(localOutfit);
    }
    saveData(appData);
    _editingOutfitId = null;

    closeModal();
    showToast(`Outfit "${name}" saved!`, 'success');

    setTimeout(() => {
      window.location.href = 'outfits.html';
    }, 1000);
  } catch (e) {
    console.error('Failed to save outfit:', e);
    showToast('Failed to save outfit to server', 'error');
  }
}

// ============================================================================
// ITEM DETAIL PAGE
// ============================================================================

async function initItemDetail() {
  const container = document.getElementById('item-detail');
  if (!container) return;

  const params = new URLSearchParams(window.location.search);
  const itemId = params.get('id');

  if (!itemId) {
    showToast('No item ID provided. Redirecting to wardrobe...', 'error');
    setTimeout(() => {
      window.location.href = 'wardrobe.html';
    }, 2000);
    return;
  }

  // Fetch item from API
  try {
    const response = await fetch(`http://localhost:8556/api/items/${itemId}`);
    if (!response.ok) {
      throw new Error('Item not found');
    }

    const item = await response.json();

    // Populate form
    document.getElementById('item-name').value = item.name || '';
    document.getElementById('item-description').value = item.description || '';
    document.getElementById('item-category').value = item.category || 'Uncategorized';
    document.getElementById('item-price').value = item.price || 0;
    document.getElementById('item-brand').value = item.brand || '';
    document.getElementById('item-color').value = item.color || '';
    document.getElementById('item-image-preview').src = item.image_url;

    const seasonEl = document.getElementById('item-season');
    if (seasonEl && item.season) seasonEl.value = item.season;
    const notesEl = document.getElementById('item-notes');
    if (notesEl) notesEl.value = item.notes || '';

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
    if (item.ai_analysis) {
      const analysisSection = document.getElementById('ai-analysis-section');
      const analysisText = document.getElementById('ai-analysis-text');
      if (analysisSection && analysisText) {
        analysisSection.style.display = 'block';
        analysisText.textContent = item.ai_analysis;
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
        enhanceBrandBtn.textContent = '⏳ Reading text with AI...';

        try {
          // Convert image URL to base64 if needed
          let imageData = item.image_url;
          if (!imageData.startsWith('data:')) {
            // Fetch image and convert to base64
            const imgResponse = await fetch(imageData);
            const blob = await imgResponse.blob();
            imageData = await new Promise((resolve) => {
              const reader = new FileReader();
              reader.onloadend = () => resolve(reader.result);
              reader.readAsDataURL(blob);
            });
          }

          // Use deepseek-ocr or qwen2-vl:7b for better text/brand detection
          const brandPrompt = `Carefully examine this clothing item image and read ALL visible text.

Focus on:
1. Clothing labels (neck tags, waistband tags, sleeve tags)
2. Printed text on the garment
3. Woven labels
4. Care labels
5. Any visible brand names, logos, or text

List every piece of text you can see in the image, especially brand names. Pay close attention to labels and tags.

Respond with JSON:
{"brand": "BRAND_NAME_HERE", "all_text": ["text1", "text2", "text3"], "description": "brief description", "category": "category", "color": "color", "tags": ["tag1", "tag2"]}

If you see a brand name, put it in the "brand" field. List ALL visible text in "all_text".`;

          // Try deepseek-ocr first (best for text), fallback to qwen2-vl:7b, then llava:13b
          let modelToUse = 'deepseek-ocr:latest';
          console.log('Attempting brand detection with:', modelToUse);

          const aiResult = await analyzeImageWithAI(imageData, modelToUse, brandPrompt);

          console.log('Brand detection result:', aiResult);

          if (aiResult) {
            // Log all detected text for debugging
            if (aiResult.all_text) {
              console.log('All detected text:', aiResult.all_text);
            }

            if (aiResult.brand) {
              document.getElementById('item-brand').value = aiResult.brand;
              const allTextInfo = aiResult.all_text ? ` (Found text: ${aiResult.all_text.join(', ')})` : '';
              showToast(`Brand detected: ${aiResult.brand}${allTextInfo}`, 'success');

              // Update via API
              await fetch(`http://localhost:8556/api/items/${itemId}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ name: item.name, category: item.category, price: item.price, brand: aiResult.brand, color: item.color })
              });
            } else {
              const textFound = aiResult.all_text && aiResult.all_text.length > 0
                ? `Text found: ${aiResult.all_text.join(', ')}. `
                : '';
              showToast(`${textFound}Could not detect brand name. Check console for details.`, 'error');
            }
          } else {
            showToast('AI analysis failed. Try manually entering brand.', 'error');
          }
        } catch (error) {
          console.error('Brand detection error:', error);
          showToast('Error detecting brand: ' + error.message, 'error');
        }

        enhanceBrandBtn.disabled = false;
        enhanceBrandBtn.textContent = '✨ Detect with AI (OCR)';
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

  // Setup retry AI analysis button
  const retryAIBtn = document.getElementById('retry-ai-btn');
  if (retryAIBtn) {
    retryAIBtn.onclick = async () => {
      if (!aiAvailable) {
        showToast('AI not available. Start the API server.', 'error');
        return;
      }
      
      retryAIBtn.disabled = true;
      retryAIBtn.textContent = '⏳ Analyzing...';
      
      await retryAIAnalysis(itemId);
      
      retryAIBtn.disabled = false;
      retryAIBtn.textContent = '🔄 Retry AI Analysis';
    };
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

      const item = appData.items.find(i => i.id == itemId);
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
        .filter(i => i.id != itemId)
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
  } catch (error) {
    console.error('Load item error:', error);
    showToast('Item not found. Redirecting to wardrobe...', 'error');
    setTimeout(() => {
      window.location.href = 'wardrobe.html';
    }, 2000);
  }
}

function displayStyleMatches(matchingItems) {
  const section = document.getElementById('style-matches-section');
  const grid = document.getElementById('style-matches-grid');

  if (!section || !grid) return;

  grid.innerHTML = '';

  matchingItems.forEach(match => {
    const item = appData.items.find(i => i.id == match.id);
    if (!item) return;

    const card = document.createElement('div');
    card.className = 'item-card';
    card.onclick = () => window.location.href = `item-detail.html?id=${item.id}`;

    const scoreColor = match.score >= 80 ? '#27AE60' :
                       match.score >= 60 ? '#F39C12' : '#E74C3C';

    card.innerHTML = `
      <img src="${item.image}" alt="${item.name}" style="width: 100%; max-height: 150px; object-fit: contain; background: var(--warm-bg); border-radius: 0.5rem;" loading="lazy">
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

async function saveItem(itemId) {
  // Fetch current item to preserve fields not in the form (ai_analysis, tags, etc.)
  let existing = {};
  try {
    const res = await fetch(`http://localhost:8556/api/items/${itemId}`);
    if (res.ok) existing = await res.json();
  } catch (e) {
    console.warn('Could not fetch existing item, some fields may be lost:', e);
  }

  const updatedItem = {
    name: document.getElementById('item-name').value,
    description: document.getElementById('item-description').value,
    category: document.getElementById('item-category').value,
    price: parseFloat(document.getElementById('item-price').value) || 0,
    brand: document.getElementById('item-brand').value,
    color: document.getElementById('item-color').value,
    season: document.getElementById('item-season')?.value || existing.season || '',
    notes: document.getElementById('item-notes')?.value || existing.notes || '',
    ai_analysis: existing.ai_analysis || '',
    tags: existing.tags || [],
  };

  try {
    const response = await fetch(`http://localhost:8556/api/items/${itemId}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(updatedItem)
    });

    if (!response.ok) {
      throw new Error('Failed to update item');
    }

    showToast('Item saved!', 'success');

    setTimeout(() => {
      window.location.href = 'wardrobe.html';
    }, 1000);
  } catch (error) {
    console.error('Save error:', error);
    showToast('Error saving item: ' + error.message, 'error');
  }
}

async function deleteItem(itemId) {
  if (!confirm('Are you sure you want to delete this item? This cannot be undone.')) {
    return;
  }

  try {
    const response = await fetch(`http://localhost:8556/api/items/${itemId}`, {
      method: 'DELETE'
    });

    if (!response.ok) {
      throw new Error('Failed to delete item');
    }

    showToast('Item deleted', 'success');

    setTimeout(() => {
      window.location.href = 'wardrobe.html';
    }, 500);
  } catch (error) {
    console.error('Delete error:', error);
    showToast('Error deleting item: ' + error.message, 'error');
  }
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

function renderOutfits(containerId, filters = {}) {
  const container = document.getElementById(containerId);
  if (!container) return;

  container.innerHTML = '';

  let outfits = [...appData.outfits];

  // Apply search filter
  if (filters.search) {
    const searchLower = filters.search.toLowerCase();
    outfits = outfits.filter(o =>
      o.name.toLowerCase().includes(searchLower) ||
      (o.notes && o.notes.toLowerCase().includes(searchLower))
    );
  }

  // Apply rating filter
  if (filters.rating && filters.rating !== 'all') {
    const minRating = parseInt(filters.rating);
    outfits = outfits.filter(o => (o.rating || 0) >= minRating);
  }

  // Apply sorting
  if (filters.sort) {
    switch(filters.sort) {
      case 'name':
        outfits.sort((a, b) => a.name.localeCompare(b.name));
        break;
      case 'rating':
        outfits.sort((a, b) => (b.rating || 0) - (a.rating || 0));
        break;
      case 'worn':
        outfits.sort((a, b) => (b.wearCount || 0) - (a.wearCount || 0));
        break;
      case 'recent':
      default:
        outfits.sort((a, b) => new Date(b.createdAt || 0) - new Date(a.createdAt || 0));
    }
  }

  if (outfits.length === 0) {
    container.innerHTML = `
      <div style="grid-column: 1/-1; text-align: center; padding: 3rem; color: var(--text-neutral);">
        <div style="font-size: 3rem; margin-bottom: 1rem;">👗</div>
        <p>No outfits found. <a href="outfit-builder.html" style="color: var(--primary);">Create your first outfit!</a></p>
      </div>
    `;
    return;
  }

  outfits.forEach(outfit => {
    const items = outfit.itemIds.map(id => appData.items.find(i => i.id == id)).filter(Boolean);

    const card = document.createElement('div');
    card.className = 'outfit-card';

    card.innerHTML = `
      <div class="outfit-preview" onclick="viewOutfitDetail(${outfit.id})">
        ${items.map((item, i) => `
          <img src="${item.image}" alt="${item.name}" style="position: absolute; left: ${50 + i * 25}px; top: ${20 + i * 30}px; max-width: ${items.length > 3 ? '80' : '100'}px; max-height: ${items.length > 3 ? '80' : '100'}px;">
        `).join('')}
      </div>
      <div class="outfit-info">
        <h3 onclick="viewOutfitDetail(${outfit.id})" style="cursor: pointer;">${outfit.name}</h3>
        <p class="outfit-meta">${items.length} items • Worn ${outfit.wearCount || 0} times</p>
        <div class="outfit-rating">
          ${'★'.repeat(Math.floor(outfit.rating || 0))}${'☆'.repeat(5 - Math.floor(outfit.rating || 0))}
          <span>${outfit.rating || 'Not rated'}</span>
        </div>
        <div style="display: flex; gap: 0.5rem; margin-top: 0.75rem; flex-wrap: wrap;">
          <button class="btn btn-secondary" style="flex: 1; padding: 0.5rem; font-size: 0.875rem;" onclick="editOutfit(${outfit.id}); event.stopPropagation();">
            ✏️ Edit
          </button>
          <button class="btn btn-secondary" style="flex: 1; padding: 0.5rem; font-size: 0.875rem;" onclick="duplicateOutfit(${outfit.id}); event.stopPropagation();">
            📋 Duplicate
          </button>
          <button class="btn btn-secondary" style="flex: 1; padding: 0.5rem; font-size: 0.875rem;" onclick="shareOutfit(${outfit.id}); event.stopPropagation();">
            🔗 Share
          </button>
          <button class="btn" style="background: #E74C3C; color: white; padding: 0.5rem; font-size: 0.875rem;" onclick="deleteOutfit(${outfit.id}); event.stopPropagation();">
            🗑️
          </button>
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

function initOutfitsFilters() {
  const searchInput = document.getElementById('outfit-search');
  const ratingFilter = document.getElementById('rating-filter');
  const sortFilter = document.getElementById('sort-filter');

  const applyFilters = () => {
    renderOutfits('outfits-grid', {
      search: searchInput?.value || '',
      rating: ratingFilter?.value || 'all',
      sort: sortFilter?.value || 'recent'
    });
  };

  if (searchInput) {
    searchInput.addEventListener('input', applyFilters);
  }
  if (ratingFilter) {
    ratingFilter.addEventListener('change', applyFilters);
  }
  if (sortFilter) {
    sortFilter.addEventListener('change', applyFilters);
  }
}

async function duplicateOutfit(outfitId) {
  const outfit = appData.outfits.find(o => o.id == outfitId);
  if (!outfit) return;

  const bodyUrl = outfit.bodyBackground ? outfit.bodyBackground.replace(/^url\(["']?/, '').replace(/["']?\)$/, '') : '';
  const positions = (outfit.positions || outfit.itemIds?.map(id => ({ id, x: 100, y: 100, z: 0 })) || []);

  try {
    const resp = await fetch('http://localhost:8556/api/outfits', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        name: `${outfit.name} (Copy)`,
        notes: outfit.notes || '',
        body_image_url: bodyUrl,
        positions: positions,
      })
    });
    if (resp.ok) {
      const result = await resp.json();
      const newOutfit = {
        ...outfit,
        id: result.id,
        name: `${outfit.name} (Copy)`,
        wearCount: 0,
        createdAt: new Date().toISOString()
      };
      appData.outfits.push(newOutfit);
      saveData(appData);
      renderOutfits('outfits-grid');
      showToast('Outfit duplicated!', 'success');
    }
  } catch (e) {
    showToast('Failed to duplicate outfit', 'error');
  }
}

function deleteOutfit(outfitId) {
  showModal('Delete Outfit?', `
    <p>Are you sure you want to delete this outfit? This action cannot be undone.</p>
  `, [
    { label: 'Cancel', onclick: 'closeModal()' },
    { label: 'Delete', primary: true, onclick: `confirmDeleteOutfit(${outfitId})` }
  ]);
}

async function confirmDeleteOutfit(outfitId) {
  try {
    await fetch(`http://localhost:8556/api/outfits/${outfitId}`, { method: 'DELETE' });
  } catch (e) {
    console.log('Failed to delete outfit from API:', e.message);
  }
  appData.outfits = appData.outfits.filter(o => o.id != outfitId);
  saveData(appData);
  closeModal();
  renderOutfits('outfits-grid');
  showToast('Outfit deleted', 'info');
}

function viewOutfitDetail(outfitId) {
  const outfit = appData.outfits.find(o => o.id === outfitId);
  if (!outfit) return;

  const items = outfit.itemIds.map(id => appData.items.find(i => i.id == id)).filter(Boolean);

  const itemsHTML = items.map(item => `
    <div style="display: inline-block; margin: 0.5rem; text-align: center;">
      <img src="${item.image}" alt="${item.name}" style="width: 80px; height: 80px; object-fit: cover; border-radius: 0.5rem;">
      <div style="font-size: 0.75rem; margin-top: 0.25rem;">${item.name}</div>
    </div>
  `).join('');

  showModal(outfit.name, `
    <div style="margin-bottom: 1rem;">
      <div class="outfit-rating" style="font-size: 1.5rem; margin-bottom: 0.5rem;">
        ${'★'.repeat(Math.floor(outfit.rating || 0))}${'☆'.repeat(5 - Math.floor(outfit.rating || 0))}
      </div>
      <p style="color: var(--text-neutral);">${items.length} items • Worn ${outfit.wearCount || 0} times</p>
      ${outfit.notes ? `<p style="margin-top: 1rem; padding: 1rem; background: var(--warm-bg); border-radius: 0.5rem;">${outfit.notes}</p>` : ''}
    </div>
    <h4>Items in this outfit:</h4>
    <div style="margin-top: 1rem;">
      ${itemsHTML}
    </div>
  `, [
    { label: 'Close', onclick: 'closeModal()' },
    { label: '👔 Worn Today', onclick: `logWearEvent(${outfit.id})` },
    { label: 'Edit', primary: true, onclick: `editOutfit(${outfit.id}); closeModal();` }
  ]);
}

async function logWearEvent(outfitId, date) {
  try {
    const resp = await fetch('http://localhost:8556/api/wear', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        outfit_id: outfitId,
        date: date || new Date().toISOString().split('T')[0],
      })
    });
    if (resp.ok) {
      const result = await resp.json();
      // Update local wear count
      const outfit = appData.outfits.find(o => o.id == outfitId);
      if (outfit) {
        outfit.wearCount = (outfit.wearCount || 0) + 1;
        // Also update item wear counts locally
        if (outfit.itemIds) {
          outfit.itemIds.forEach(itemId => {
            const item = appData.items.find(i => i.id == itemId);
            if (item) item.wear_count = (item.wear_count || 0) + 1;
          });
        }
        saveData(appData);
      }
      closeModal();
      renderOutfits('outfits-grid');
      showToast(`Logged wear for ${result.items_worn} items!`, 'success');
    } else {
      showToast('Failed to log wear event', 'error');
    }
  } catch (e) {
    console.error('Failed to log wear:', e);
    showToast('Failed to log wear event', 'error');
  }
}

function editOutfit(outfitId) {
  // Load outfit into builder
  const outfit = appData.outfits.find(o => o.id == outfitId);
  if (!outfit) return;

  sessionStorage.setItem('editOutfitId', outfitId.toString());
  window.location.href = 'outfit-builder.html';
}

// ============================================================================
// ANALYTICS
// ============================================================================

function initAnalytics() {
  renderAnalytics();
  renderMostWornItems();
  renderLeastWornItems();
  renderCategoryBreakdown();

  // Setup AI Analytics button
  const aiAnalyticsBtn = document.getElementById('ai-analytics-btn');
  if (aiAnalyticsBtn) {
    aiAnalyticsBtn.onclick = async () => {
      if (!aiAvailable) {
        showToast('AI not available. Start the API server.', 'error');
        return;
      }

      aiAnalyticsBtn.disabled = true;
      aiAnalyticsBtn.textContent = '⏳ Analyzing...';

      const insights = await getAIAnalyticsInsights();

      aiAnalyticsBtn.disabled = false;
      aiAnalyticsBtn.textContent = '✨ Get AI Insights';

      if (insights) {
        displayAIAnalyticsInsights(insights);
      }
    };
  }
}

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

function calculateCostPerWear(item) {
  if (!item.price) return 0;
  if (item.wearCount > 0) {
    return item.price / item.wearCount;
  }
  return item.price;
}

function getMostWornItems(limit = 5) {
  return [...appData.items]
    .filter(item => item.wearCount > 0)
    .sort((a, b) => b.wearCount - a.wearCount)
    .slice(0, limit);
}

function getLeastWornItems(limit = 5) {
  return [...appData.items]
    .sort((a, b) => (a.wearCount || 0) - (b.wearCount || 0))
    .slice(0, limit);
}

function renderMostWornItems() {
  const tbody = document.getElementById('most-worn-tbody');
  if (!tbody) return;

  const items = getMostWornItems(5);
  tbody.innerHTML = items.map(item => {
    const cpw = calculateCostPerWear(item);
    return `
      <tr style="border-bottom: 1px solid var(--border);">
        <td style="padding: 1rem; font-weight: 600;">${item.name}</td>
        <td style="padding: 1rem; color: var(--text-neutral);">${item.category}</td>
        <td style="padding: 1rem;">${item.wearCount || 0}</td>
        <td style="padding: 1rem; color: var(--primary);">$${cpw.toFixed(2)}</td>
      </tr>
    `;
  }).join('');
}

function renderLeastWornItems() {
  const tbody = document.getElementById('least-worn-tbody');
  if (!tbody) return;

  const items = getLeastWornItems(5);
  tbody.innerHTML = items.map(item => {
    const cpw = calculateCostPerWear(item);
    const color = item.wearCount < 3 ? '#E74C3C' : 'var(--primary)';
    return `
      <tr style="border-bottom: 1px solid var(--border);">
        <td style="padding: 1rem; font-weight: 600;">${item.name}</td>
        <td style="padding: 1rem; color: var(--text-neutral);">${item.category}</td>
        <td style="padding: 1rem;">${item.wearCount || 0}</td>
        <td style="padding: 1rem; color: ${color};">$${cpw.toFixed(2)}</td>
      </tr>
    `;
  }).join('');
}

function renderCategoryBreakdown() {
  const container = document.getElementById('category-breakdown');
  if (!container) return;

  const categoryCounts = {};
  appData.items.forEach(item => {
    categoryCounts[item.category] = (categoryCounts[item.category] || 0) + 1;
  });

  const total = appData.items.length;
  const breakdown = Object.entries(categoryCounts)
    .map(([category, count]) => ({
      category,
      count,
      percentage: ((count / total) * 100).toFixed(1)
    }))
    .sort((a, b) => b.count - a.count);

  container.innerHTML = `
    <div style="text-align: center;">
      ${breakdown.map(item => `
        <div style="margin: 1rem 0;">
          <strong>${item.category}:</strong> ${item.count} items (${item.percentage}%)
          <div style="background: var(--border); height: 8px; border-radius: 4px; margin-top: 0.5rem;">
            <div style="background: var(--primary); height: 8px; border-radius: 4px; width: ${item.percentage}%;"></div>
          </div>
        </div>
      `).join('')}
    </div>
  `;
}

async function getAIAnalyticsInsights() {
  if (!aiAvailable) return null;

  try {
    const wardrobeData = {
      items: appData.items.map(item => ({
        id: item.id,
        name: item.name,
        category: item.category,
        color: item.color,
        brand: item.brand,
        price: item.price,
        wearCount: item.wearCount || 0
      })),
      totalInvestment: appData.items.reduce((sum, item) => sum + (item.price || 0), 0),
      totalItems: appData.items.length
    };

    const resp = await fetch(`${AI_API_URL}/analytics-insights`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(wardrobeData)
    });

    if (resp.ok) {
      return await resp.json();
    } else {
      const error = await resp.json();
      console.error('Analytics insights error:', error);
      showToast(`AI error: ${error.error || 'Unknown error'}`, 'error');
    }
  } catch (e) {
    console.error('Analytics insights failed:', e);
    showToast('Analytics insights failed: ' + e.message, 'error');
  }
  return null;
}

function displayAIAnalyticsInsights(insights) {
  const container = document.getElementById('ai-insights-container');
  if (!container) {
    showModal('AI Analytics Insights', `
      <div style="padding: 1rem;">
        <h3 style="margin-top: 0;">Key Insights</h3>
        <p style="line-height: 1.8; color: var(--text-neutral);">${insights.summary || insights.insights || 'No insights available'}</p>
        ${insights.recommendations ? `
          <h3>Recommendations</h3>
          <ul style="line-height: 2;">
            ${(Array.isArray(insights.recommendations) ? insights.recommendations : [insights.recommendations]).map(r => `<li>${r}</li>`).join('')}
          </ul>
        ` : ''}
      </div>
    `, [{ label: 'Close', onclick: 'closeModal()' }]);
  } else {
    container.innerHTML = `
      <div class="card" style="background: var(--warm-bg);">
        <h3 style="margin-top: 0; color: var(--primary);">✨ AI Insights</h3>
        <p style="line-height: 1.8; color: var(--text-neutral);">${insights.summary || insights.insights || 'No insights available'}</p>
        ${insights.recommendations ? `
          <h4>Recommendations</h4>
          <ul style="line-height: 2;">
            ${(Array.isArray(insights.recommendations) ? insights.recommendations : [insights.recommendations]).map(r => `<li>${r}</li>`).join('')}
          </ul>
        ` : ''}
      </div>
    `;
    container.style.display = 'block';
    container.scrollIntoView({ behavior: 'smooth' });
  }
  showToast('AI insights generated!', 'success');
}

// ============================================================================
// CALENDAR PLANNING
// ============================================================================

let currentCalendarDate = new Date();

function initCalendar() {
  renderCalendar();

  // Setup AI Suggestions button
  const aiSuggestBtn = document.getElementById('ai-calendar-suggest-btn');
  if (aiSuggestBtn) {
    aiSuggestBtn.onclick = () => {
      const today = new Date();
      const dateStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;
      openDayModal(today.getFullYear(), today.getMonth(), today.getDate());
    };
  }
}

function renderCalendar() {
  const year = currentCalendarDate.getFullYear();
  const month = currentCalendarDate.getMonth();

  // Update month header
  const monthHeader = document.getElementById('current-month');
  if (monthHeader) {
    const monthNames = ['January', 'February', 'March', 'April', 'May', 'June',
                        'July', 'August', 'September', 'October', 'November', 'December'];
    monthHeader.textContent = `${monthNames[month]} ${year}`;
  }

  // Get calendar container
  const container = document.getElementById('calendar-days');
  if (!container) return;

  container.innerHTML = '';

  // Get first day of month and total days
  const firstDay = new Date(year, month, 1).getDay();
  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const prevMonthDays = new Date(year, month, 0).getDate();

  // Today's date for highlighting
  const today = new Date();
  const isCurrentMonth = today.getFullYear() === year && today.getMonth() === month;

  // Add previous month's trailing days
  for (let i = firstDay - 1; i >= 0; i--) {
    const day = prevMonthDays - i;
    const dayEl = createCalendarDay(year, month - 1, day, true);
    container.appendChild(dayEl);
  }

  // Add current month's days
  for (let day = 1; day <= daysInMonth; day++) {
    const isToday = isCurrentMonth && day === today.getDate();
    const dayEl = createCalendarDay(year, month, day, false, isToday);
    container.appendChild(dayEl);
  }

  // Add next month's leading days to fill grid
  const totalCells = container.children.length;
  const remainingCells = 42 - totalCells; // 6 weeks * 7 days
  for (let day = 1; day <= remainingCells; day++) {
    const dayEl = createCalendarDay(year, month + 1, day, true);
    container.appendChild(dayEl);
  }
}

function createCalendarDay(year, month, day, isOtherMonth, isToday = false) {
  const dateStr = `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
  const event = appData.calendarEvents[dateStr];

  const dayEl = document.createElement('div');
  dayEl.className = `calendar-day ${isOtherMonth ? 'other-month' : ''} ${isToday ? 'today' : ''} ${event ? 'has-outfit' : ''}`;
  dayEl.onclick = () => openDayModal(year, month, day);

  let outfitHTML = '';
  let eventHTML = '';

  if (event) {
    if (event.outfitId) {
      const outfit = appData.outfits.find(o => o.id == event.outfitId);
      if (outfit) {
        const items = outfit.itemIds.map(id => appData.items.find(i => i.id == id)).filter(Boolean).slice(0, 3);
        outfitHTML = `
          <div class="day-outfit">
            ${items.map(item => `<img src="${item.image}" class="day-outfit-thumb" alt="${item.name}" title="${item.name}">`).join('')}
          </div>
        `;
      }
    }

    if (event.eventName) {
      eventHTML = `<div class="day-event">${event.eventName}</div>`;
    }
  }

  dayEl.innerHTML = `
    <span class="day-number">${day}</span>
    ${event && event.weather ? `<span class="weather-badge">${event.weather}</span>` : ''}
    ${outfitHTML}
    ${eventHTML}
  `;

  return dayEl;
}

function changeMonth(delta) {
  currentCalendarDate.setMonth(currentCalendarDate.getMonth() + delta);
  renderCalendar();
}

function openDayModal(year, month, day) {
  const dateStr = `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
  const event = appData.calendarEvents[dateStr] || {};

  const outfitsOptions = appData.outfits.map(outfit =>
    `<option value="${outfit.id}" ${event.outfitId == outfit.id ? 'selected' : ''}>${outfit.name}</option>`
  ).join('');

  // Build buttons array
  const buttons = [{ label: 'Cancel', onclick: 'closeModal()' }];

  // Add "View Outfit" button if outfit is selected
  if (event.outfitId) {
    buttons.push({
      label: 'View Outfit',
      onclick: `closeModal(); viewOutfitFromCalendar(${event.outfitId})`
    });
  }

  buttons.push({ label: 'Save', primary: true, onclick: `saveCalendarOutfit('${dateStr}')` });

  showModal(`Plan Outfit for ${month + 1}/${day}/${year}`, `
    <div class="form-group">
      <label>Select Outfit</label>
      <select id="calendar-outfit-select" onchange="updateViewOutfitButton()">
        <option value="">No outfit planned</option>
        ${outfitsOptions}
      </select>
    </div>
    <div class="form-group">
      <label>Event/Occasion</label>
      <input type="text" id="calendar-event-input" value="${event.eventName || ''}" placeholder="e.g., Work meeting, Date night">
    </div>
    <div class="form-group">
      <label>Location</label>
      <input type="text" id="calendar-location-input" value="${event.location || ''}" placeholder="e.g., Office, Restaurant, Park">
    </div>
    <div class="form-group">
      <label>Weather</label>
      <select id="calendar-weather-select">
        <option value="">Unknown</option>
        <option value="☀️" ${event.weather === '☀️' ? 'selected' : ''}>☀️ Sunny</option>
        <option value="🌤️" ${event.weather === '🌤️' ? 'selected' : ''}>🌤️ Partly Cloudy</option>
        <option value="⛅" ${event.weather === '⛅' ? 'selected' : ''}>⛅ Cloudy</option>
        <option value="🌧️" ${event.weather === '🌧️' ? 'selected' : ''}>🌧️ Rainy</option>
        <option value="❄️" ${event.weather === '❄️' ? 'selected' : ''}>❄️ Snowy</option>
      </select>
    </div>
    ${aiAvailable ? `
      <button class="btn btn-secondary" style="width: 100%; margin-top: 1rem;" onclick="getAICalendarSuggestionsForDay('${dateStr}')">
        ✨ Get AI Outfit Suggestions
      </button>
    ` : ''}
  `, buttons);
}

function viewOutfitFromCalendar(outfitId) {
  // Navigate to outfits page with the outfit ID
  window.location.href = `outfits.html?outfit=${outfitId}`;
}

async function saveCalendarOutfit(dateStr) {
  const outfitSelect = document.getElementById('calendar-outfit-select');
  const eventInput = document.getElementById('calendar-event-input');
  const locationInput = document.getElementById('calendar-location-input');
  const weatherSelect = document.getElementById('calendar-weather-select');

  const outfitId = outfitSelect?.value ? parseInt(outfitSelect.value) : null;
  const eventName = eventInput?.value.trim() || '';
  const location = locationInput?.value.trim() || '';
  const weather = weatherSelect?.value || '';

  if (outfitId || eventName || weather || location) {
    // Save to API
    try {
      await fetch('http://localhost:8556/api/calendar', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          date: dateStr,
          event_name: eventName,
          weather: weather,
          location: location,
          outfit_id: outfitId,
        })
      });
    } catch (e) {
      console.log('Failed to save calendar event to API:', e.message);
    }

    appData.calendarEvents[dateStr] = {
      outfitId,
      eventName,
      location,
      weather,
      date: dateStr
    };
  } else {
    // Delete from API
    try {
      await fetch(`http://localhost:8556/api/calendar/${dateStr}`, { method: 'DELETE' });
    } catch (e) {
      console.log('Failed to delete calendar event from API:', e.message);
    }
    delete appData.calendarEvents[dateStr];
  }

  saveData(appData);
  closeModal();
  renderCalendar();
  showToast('Calendar event saved!', 'success');
}

async function getAICalendarSuggestionsForDay(dateStr) {
  const eventInput = document.getElementById('calendar-event-input');
  const weatherSelect = document.getElementById('calendar-weather-select');

  const eventName = eventInput?.value.trim() || 'casual day';
  const weather = weatherSelect?.value || '☀️';

  const suggestions = await getAICalendarSuggestions(dateStr, eventName, weather);

  if (suggestions && suggestions.suggested_outfits && suggestions.suggested_outfits.length > 0) {
    displayCalendarSuggestions(dateStr, suggestions);
  } else {
    showToast('Could not get AI suggestions', 'error');
  }
}

async function getAICalendarSuggestions(date, event, weather = 'sunny') {
  if (!aiAvailable) return null;

  try {
    const items = appData.items.map(item => ({
      id: item.id,
      name: item.name,
      category: item.category,
      color: item.color,
      tags: item.tags || []
    }));

    const resp = await fetch(`${AI_API_URL}/calendar-suggestions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ date, event, weather, items })
    });

    if (resp.ok) {
      return await resp.json();
    } else {
      const error = await resp.json();
      console.error('Calendar suggestions error:', error);
      showToast(`AI error: ${error.error || 'Unknown error'}`, 'error');
    }
  } catch (e) {
    console.error('Calendar suggestions failed:', e);
    showToast('Calendar suggestions failed: ' + e.message, 'error');
  }
  return null;
}

function displayCalendarSuggestions(dateStr, suggestions) {
  const outfitsHTML = suggestions.suggested_outfits.map(outfit => {
    const items = outfit.item_ids.map(id => appData.items.find(i => i.id === id)).filter(Boolean);
    return `
      <div style="padding: 1rem; background: var(--warm-bg); border-radius: 0.5rem; margin-bottom: 1rem;">
        <h4 style="margin-top: 0;">${outfit.name || 'Suggested Outfit'}</h4>
        <p style="color: var(--text-neutral); font-size: 0.875rem;">${outfit.reason || ''}</p>
        <div style="display: flex; gap: 0.5rem; margin-top: 0.5rem;">
          ${items.map(item => `<img src="${item.image}" style="width: 60px; height: 60px; object-fit: cover; border-radius: 0.25rem;" alt="${item.name}">`).join('')}
        </div>
        <button class="btn btn-primary" style="margin-top: 0.5rem; width: 100%;" onclick="applySuggestedOutfitToCalendar('${dateStr}', ${JSON.stringify(outfit.item_ids).replace(/"/g, '&quot;')})">
          Use This Outfit
        </button>
      </div>
    `;
  }).join('');

  showModal('AI Outfit Suggestions', `
    <div>
      <p style="color: var(--text-neutral); margin-bottom: 1rem;">${suggestions.reasoning || 'Here are some outfit suggestions for this day:'}</p>
      ${outfitsHTML}
    </div>
  `, [
    { label: 'Close', onclick: 'closeModal()' }
  ]);
}

window.applySuggestedOutfitToCalendar = function(dateStr, itemIds) {
  // Create a new outfit from suggested items
  const newOutfit = {
    id: Date.now(),
    name: `Calendar Outfit ${dateStr}`,
    itemIds: itemIds,
    positions: itemIds.map((id, i) => ({ id, x: 100 + i * 30, y: 100 + i * 40, z: i })),
    wearCount: 0,
    rating: 0,
    createdAt: new Date().toISOString()
  };

  appData.outfits.push(newOutfit);

  appData.calendarEvents[dateStr] = {
    ...appData.calendarEvents[dateStr],
    outfitId: newOutfit.id
  };

  saveData(appData);
  closeModal();
  renderCalendar();
  showToast('Outfit added to calendar!', 'success');
};

// ============================================================================
// SETTINGS
// ============================================================================

const SETTINGS_KEY = 'gothreads_settings';

const defaultSettings = {
  autoRemoveBg: true,
  autoTag: true,
  aiProvider: 'local',
  currency: 'USD',
  dateFormat: 'MM/DD/YYYY',
  maxWearsBeforeLaundry: 3,
  notifications: true,
  emailSummaries: false
};

function getSettings() {
  const stored = localStorage.getItem(SETTINGS_KEY);
  if (stored) {
    return { ...defaultSettings, ...JSON.parse(stored) };
  }
  return defaultSettings;
}

function initSettings() {
  const settings = getSettings();

  // Populate form fields
  const autoRemoveBgInput = document.getElementById('setting-auto-remove-bg');
  const autoTagInput = document.getElementById('setting-auto-tag');
  const aiProviderSelect = document.getElementById('setting-ai-provider');
  const currencySelect = document.getElementById('setting-currency');
  const dateFormatSelect = document.getElementById('setting-date-format');
  const maxWearsInput = document.getElementById('setting-max-wears');
  const notificationsInput = document.getElementById('setting-notifications');
  const emailSummariesInput = document.getElementById('setting-email-summaries');

  if (autoRemoveBgInput) autoRemoveBgInput.checked = settings.autoRemoveBg;
  if (autoTagInput) autoTagInput.checked = settings.autoTag;
  if (aiProviderSelect) aiProviderSelect.value = settings.aiProvider;
  if (currencySelect) currencySelect.value = settings.currency;
  if (dateFormatSelect) dateFormatSelect.value = settings.dateFormat;
  if (maxWearsInput) maxWearsInput.value = settings.maxWearsBeforeLaundry;
  if (notificationsInput) notificationsInput.checked = settings.notifications;
  if (emailSummariesInput) emailSummariesInput.checked = settings.emailSummaries;

  // Setup save button
  const saveBtn = document.getElementById('save-settings-btn');
  if (saveBtn) {
    saveBtn.onclick = saveSettings;
  }

  // Setup export buttons
  const exportJsonBtn = document.getElementById('export-json-btn');
  const exportCsvBtn = document.getElementById('export-csv-btn');
  const exportPdfBtn = document.getElementById('export-pdf-btn');

  if (exportJsonBtn) exportJsonBtn.onclick = () => exportData('json');
  if (exportCsvBtn) exportCsvBtn.onclick = () => exportData('csv');
  if (exportPdfBtn) exportPdfBtn.onclick = () => exportData('pdf');

  // Setup import button
  const importBtn = document.getElementById('import-data-btn');
  const importInput = document.getElementById('import-file-input');
  if (importBtn && importInput) {
    importBtn.onclick = () => importInput.click();
    importInput.onchange = (e) => {
      const file = e.target.files[0];
      if (file) {
        const reader = new FileReader();
        reader.onload = (event) => {
          try {
            const data = JSON.parse(event.target.result);
            importData(data);
          } catch (error) {
            showToast('Invalid JSON file', 'error');
          }
        };
        reader.readAsText(file);
      }
    };
  }
}

function saveSettings() {
  const settings = {
    autoRemoveBg: document.getElementById('setting-auto-remove-bg')?.checked ?? true,
    autoTag: document.getElementById('setting-auto-tag')?.checked ?? true,
    aiProvider: document.getElementById('setting-ai-provider')?.value || 'local',
    currency: document.getElementById('setting-currency')?.value || 'USD',
    dateFormat: document.getElementById('setting-date-format')?.value || 'MM/DD/YYYY',
    maxWearsBeforeLaundry: parseInt(document.getElementById('setting-max-wears')?.value) || 3,
    notifications: document.getElementById('setting-notifications')?.checked ?? true,
    emailSummaries: document.getElementById('setting-email-summaries')?.checked ?? false
  };

  localStorage.setItem(SETTINGS_KEY, JSON.stringify(settings));
  showToast('Settings saved!', 'success');
}

function exportData(format) {
  const data = {
    items: appData.items,
    outfits: appData.outfits,
    wearHistory: appData.wearHistory,
    calendarEvents: appData.calendarEvents,
    exportedAt: new Date().toISOString()
  };

  if (format === 'json') {
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    downloadFile(blob, `gothreads-export-${Date.now()}.json`);
    showToast('Data exported as JSON!', 'success');
  } else if (format === 'csv') {
    const csv = convertToCSV(data.items);
    const blob = new Blob([csv], { type: 'text/csv' });
    downloadFile(blob, `gothreads-items-${Date.now()}.csv`);
    showToast('Items exported as CSV!', 'success');
  } else if (format === 'pdf') {
    showToast('PDF export would generate a report (not implemented in mockup)', 'info');
  }
}

function convertToCSV(items) {
  const headers = ['ID', 'Name', 'Category', 'Price', 'Color', 'Brand', 'Wear Count'];
  const rows = items.map(item => [
    item.id,
    item.name,
    item.category,
    item.price,
    item.color,
    item.brand,
    item.wearCount || 0
  ]);

  return [headers.join(','), ...rows.map(row => row.join(','))].join('\n');
}

function downloadFile(blob, filename) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

function importData(jsonData) {
  if (!jsonData.items || !Array.isArray(jsonData.items)) {
    showToast('Invalid data format', 'error');
    return;
  }

  showModal('Import Data?', `
    <p>This will import ${jsonData.items.length} items${jsonData.outfits ? ` and ${jsonData.outfits.length} outfits` : ''}.</p>
    <p style="color: #F39C12;">Existing data will be merged (not replaced).</p>
  `, [
    { label: 'Cancel', onclick: 'closeModal()' },
    { label: 'Import', primary: true, onclick: `confirmImport(${JSON.stringify(jsonData).replace(/"/g, '&quot;')})` }
  ]);
}

window.confirmImport = function(data) {
  // Merge items
  const existingIds = new Set(appData.items.map(i => i.id));
  const newItems = data.items.filter(item => !existingIds.has(item.id));
  appData.items.push(...newItems);

  // Merge outfits
  if (data.outfits) {
    const existingOutfitIds = new Set(appData.outfits.map(o => o.id));
    const newOutfits = data.outfits.filter(outfit => !existingOutfitIds.has(outfit.id));
    appData.outfits.push(...newOutfits);
  }

  // Merge calendar events
  if (data.calendarEvents) {
    appData.calendarEvents = { ...appData.calendarEvents, ...data.calendarEvents };
  }

  saveData(appData);
  closeModal();
  showToast(`Imported ${newItems.length} new items!`, 'success');
  setTimeout(() => window.location.reload(), 1500);
};

// ============================================================================
// SHARING SYSTEM
// ============================================================================

function shareOutfit(outfitId) {
  const outfit = appData.outfits.find(o => o.id === outfitId);
  if (!outfit) return;

  const shareId = generateShareId();
  const shareUrl = `${window.location.origin}${window.location.pathname.replace(/[^/]*$/, '')}share.html?id=${shareId}`;

  // Store shared outfit
  const sharedOutfit = {
    shareId,
    outfitId,
    outfit: { ...outfit },
    createdAt: new Date().toISOString(),
    ratings: [],
    views: 0
  };

  if (!appData.sharedOutfits) appData.sharedOutfits = [];
  appData.sharedOutfits.push(sharedOutfit);
  saveData(appData);

  showModal('Share Outfit', `
    <div style="text-align: center; padding: 1rem;">
      <h3 style="margin-top: 0;">Share "${outfit.name}"</h3>
      <p style="color: var(--text-neutral); margin-bottom: 1.5rem;">Anyone with this link can view your outfit</p>
      <div style="background: var(--warm-bg); padding: 1rem; border-radius: 0.5rem; margin-bottom: 1rem; word-break: break-all;">
        <code id="share-url">${shareUrl}</code>
      </div>
      <button class="btn btn-primary" onclick="copyToClipboard('${shareUrl}')" style="width: 100%;">
        🔗 Copy Link
      </button>
    </div>
  `, [
    { label: 'Close', onclick: 'closeModal()' }
  ]);
}

function generateShareId() {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let id = '';
  for (let i = 0; i < 8; i++) {
    id += chars.charAt(Math.floor(Math.random() * chars.length));
  }
  return id;
}

function loadSharedOutfit(shareId) {
  const shared = appData.sharedOutfits?.find(s => s.shareId === shareId);
  if (!shared) return null;

  shared.views++;
  saveData(appData);
  return shared;
}

function rateSharedOutfit(shareId, rating) {
  const shared = appData.sharedOutfits?.find(s => s.shareId === shareId);
  if (!shared) return;

  if (!shared.ratings) shared.ratings = [];
  shared.ratings.push({ rating, date: new Date().toISOString() });
  saveData(appData);

  showToast(`Rated ${rating} stars!`, 'success');
}

window.copyToClipboard = function(text) {
  navigator.clipboard.writeText(text).then(() => {
    showToast('Link copied to clipboard!', 'success');
  });
};

// ============================================================================
// CHAT INTERFACE FOR OUTFIT CREATION
// ============================================================================

let chatMessages = [];

let chatMessageHistory = []; // [{role, content}] for multi-turn conversation

function initChat() {
  const chatBtn = document.getElementById('chat-btn');
  const chatWidget = document.getElementById('chat-widget');
  const chatClose = document.getElementById('chat-close');
  const chatInput = document.getElementById('chat-input');
  const chatSend = document.getElementById('chat-send-btn');

  if (chatBtn && chatWidget) {
    chatBtn.onclick = () => {
      chatWidget.style.display = chatWidget.style.display === 'none' ? 'flex' : 'none';
    };
  }

  if (chatClose && chatWidget) {
    chatClose.onclick = () => {
      chatWidget.style.display = 'none';
    };
  }

  if (chatSend && chatInput) {
    chatSend.onclick = () => {
      const message = chatInput.value.trim();
      if (message) {
        sendChatMessage(message);
        chatInput.value = '';
      }
    };

    chatInput.onkeypress = (e) => {
      if (e.key === 'Enter') {
        chatSend.click();
      }
    };
  }
}

async function sendChatMessage(message) {
  const chatMessages = document.getElementById('chat-messages');
  if (!chatMessages) return;

  // Add user message
  const userMsg = document.createElement('div');
  userMsg.className = 'chat-message user-message';
  userMsg.textContent = message;
  chatMessages.appendChild(userMsg);
  chatMessages.scrollTop = chatMessages.scrollHeight;

  // Add loading message
  const loadingMsg = document.createElement('div');
  loadingMsg.className = 'chat-message ai-message';
  loadingMsg.innerHTML = '<div class="spinner" style="width: 20px; height: 20px;"></div> Thinking...';
  chatMessages.appendChild(loadingMsg);
  chatMessages.scrollTop = chatMessages.scrollHeight;

  try {
    // Submit chat job via async job queue
    const jobResp = await fetch(`${AI_API_URL}/ai/jobs`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        type: 'chat',
        payload: {
          message: message,
          history: chatMessageHistory,
        }
      })
    });

    if (!jobResp.ok) {
      throw new Error('Failed to submit chat job');
    }

    const { id: jobId } = await jobResp.json();

    // Listen for SSE updates
    const eventSource = new EventSource(`${AI_API_URL}/ai/jobs/${jobId}`);

    eventSource.addEventListener('status', (e) => {
      loadingMsg.innerHTML = `<div class="spinner" style="width: 20px; height: 20px; display: inline-block; vertical-align: middle;"></div> ${e.data}`;
      chatMessages.scrollTop = chatMessages.scrollHeight;
    });

    eventSource.addEventListener('result', (e) => {
      eventSource.close();
      loadingMsg.remove();

      const data = JSON.parse(e.data);

      // Store in conversation history for multi-turn
      chatMessageHistory.push({ role: 'user', content: message });
      chatMessageHistory.push({ role: 'assistant', content: data.message });

      // Keep history reasonable (last 20 messages)
      if (chatMessageHistory.length > 20) {
        chatMessageHistory = chatMessageHistory.slice(-20);
      }

      processChatResponse(data);
    });

    eventSource.addEventListener('error', (e) => {
      eventSource.close();
      loadingMsg.remove();
      const errorData = e.data || 'AI request failed';
      const errorMsg = document.createElement('div');
      errorMsg.className = 'chat-message ai-message';
      errorMsg.textContent = `Sorry: ${errorData}`;
      chatMessages.appendChild(errorMsg);
      chatMessages.scrollTop = chatMessages.scrollHeight;
    });

  } catch (e) {
    loadingMsg.remove();
    const errorMsg = document.createElement('div');
    errorMsg.className = 'chat-message ai-message';
    errorMsg.textContent = 'AI service unavailable. Make sure the API server is running.';
    chatMessages.appendChild(errorMsg);
  }

  chatMessages.scrollTop = chatMessages.scrollHeight;
}

function processChatResponse(response) {
  const chatMessages = document.getElementById('chat-messages');
  if (!chatMessages) return;

  // Add AI response
  const aiMsg = document.createElement('div');
  aiMsg.className = 'chat-message ai-message';
  aiMsg.style.whiteSpace = 'pre-wrap';
  aiMsg.textContent = response.message || 'Here are some suggestions!';
  chatMessages.appendChild(aiMsg);

  if (response.tools_used > 0) {
    const toolsNote = document.createElement('div');
    toolsNote.className = 'chat-message ai-message';
    toolsNote.style.cssText = 'font-size: 0.7rem; color: var(--text-neutral); padding: 0.25rem 0.75rem;';
    toolsNote.textContent = `📊 Queried wardrobe (${response.tools_used} lookup${response.tools_used > 1 ? 's' : ''})`;
    chatMessages.appendChild(toolsNote);
  }

  chatMessages.scrollTop = chatMessages.scrollHeight;
}

// ============================================================================
// AI ENHANCEMENTS
// ============================================================================

async function regenerateOutfit(outfitId, modifier) {
  if (!aiAvailable) {
    showToast('AI not available', 'error');
    return null;
  }

  const outfit = appData.outfits.find(o => o.id === outfitId);
  if (!outfit) return null;

  const items = outfit.itemIds.map(id => appData.items.find(i => i.id == id)).filter(Boolean);

  try {
    const resp = await fetch(`${AI_API_URL}/outfit-regenerate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        items: items.map(i => ({ id: i.id, name: i.name, category: i.category, color: i.color })),
        modifier,
        allItems: appData.items.map(i => ({ id: i.id, name: i.name, category: i.category, color: i.color }))
      })
    });

    if (resp.ok) {
      return await resp.json();
    } else {
      const error = await resp.json();
      console.error('Outfit regeneration error:', error);
      showToast(`AI error: ${error.error || 'Unknown error'}`, 'error');
    }
  } catch (e) {
    console.error('Outfit regeneration failed:', e);
    showToast('Outfit regeneration failed: ' + e.message, 'error');
  }
  return null;
}

async function getStyleVariation(items, style) {
  if (!aiAvailable) return null;

  try {
    const resp = await fetch(`${AI_API_URL}/style-variation`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ items, style })
    });

    if (resp.ok) {
      return await resp.json();
    }
  } catch (e) {
    console.error('Style variation failed:', e);
  }
  return null;
}

// ============================================================================
// INITIALIZATION
// ============================================================================

document.addEventListener('DOMContentLoaded', async function() {
  // Migrate localStorage data to API (one-time)
  await migrateLocalStorageToAPI();

  // Sync from DB into appData before rendering anything
  await syncItemsFromAPI();
  await syncOutfitsFromAPI();
  await syncCalendarFromAPI();

  // Check AI status on load
  checkAIStatus();

  // Initialize notification panel for AI job progress
  initNotificationPanel();

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
  initOutfitsFilters();

  // Analytics
  if (document.getElementById('stat-total-items')) {
    initAnalytics();
  }

  // Calendar
  if (document.getElementById('calendar-days') || document.getElementById('current-month')) {
    initCalendar();
  }

  // Settings
  if (document.getElementById('setting-auto-remove-bg')) {
    initSettings();
  }

  // Chat interface
  if (document.getElementById('chat-widget')) {
    initChat();
  }

  // Load outfit if editing
  const editOutfitId = sessionStorage.getItem('editOutfitId');
  if (editOutfitId && document.querySelector('.outfit-canvas')) {
    const outfit = appData.outfits.find(o => String(o.id) === String(editOutfitId));
    if (outfit) {
      _editingOutfitId = outfit.id; // Track for save (PUT vs POST)

      // Restore body background if it exists
      const canvas = document.querySelector('.outfit-canvas');
      if (outfit.bodyBackground && canvas) {
        canvas.style.backgroundImage = outfit.bodyBackground;
      }

      // Load items onto canvas (use == for loose type comparison)
      outfit.itemIds.forEach(itemId => {
        const item = appData.items.find(i => i.id == itemId);
        if (item) {
          const pos = outfit.positions ? outfit.positions.find(p => p.id == itemId) : null;
          addToCanvasAtPosition(item, pos?.x || 100, pos?.y || 100, pos?.z || 0);
        }
      });
      sessionStorage.removeItem('editOutfitId');
    }
  }

  // Load shared outfit if on share page
  const urlParams = new URLSearchParams(window.location.search);
  const shareId = urlParams.get('id');
  if (shareId && window.location.pathname.includes('share.html')) {
    const shared = loadSharedOutfit(shareId);
    if (shared) {
      renderSharedOutfit(shared);
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
    <img src="${item.image}" alt="${item.name}" draggable="false" style="width: 100%; height: 100%; object-fit: contain;">
    <div class="canvas-item-controls">
      <button onclick="resizeCanvasItem(this.closest('.canvas-item'), 'bigger')" title="Make bigger">+</button>
      <button onclick="resizeCanvasItem(this.closest('.canvas-item'), 'smaller')" title="Make smaller">−</button>
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

  // Hide placeholder when items are added
  const placeholder = document.getElementById('canvas-placeholder');
  if (placeholder) {
    placeholder.style.display = 'none';
  }

  // Update picker to show checkmarks
  renderPickerItems(document.querySelector('.picker-tab.active')?.dataset.category || 'all');
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
        const item = appData.items.find(i => i.id == id);
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
  let missingItemsHTML = '';
  if (result.missing_items && result.missing_items.length > 0) {
    missingItemsHTML = `
      <div style="margin-bottom: 1rem;">
        <h3 style="color: var(--primary); margin-top: 0;">🛍️ Suggested Items to Add</h3>
        <ul style="line-height: 2; padding-left: 1.5rem;">
          ${result.missing_items.map(item => `<li>${item}</li>`).join('')}
        </ul>
      </div>
    `;
  }

  let strengthsHTML = '';
  if (result.strengths && result.strengths.length > 0) {
    strengthsHTML = `
      <div style="margin-bottom: 1rem;">
        <h3 style="color: #27AE60; margin-top: 0;">✅ What You Have Covered</h3>
        <ul style="line-height: 2; padding-left: 1.5rem;">
          ${result.strengths.map(s => `<li>${s}</li>`).join('')}
        </ul>
      </div>
    `;
  }

  let tipsHTML = '';
  if (result.tips && result.tips.length > 0) {
    tipsHTML = `
      <div style="padding: 1rem; background: var(--warm-bg); border-radius: 0.5rem;">
        <h3 style="margin-top: 0;">💡 Wardrobe Tips</h3>
        <ul style="line-height: 2; padding-left: 1.5rem;">
          ${result.tips.map(tip => `<li>${tip}</li>`).join('')}
        </ul>
      </div>
    `;
  }

  showModal('👗 Wardrobe Analysis', `
    <div style="padding: 1rem; background: var(--warm-bg); border-radius: 0.5rem; margin-bottom: 1rem;">
      <h3 style="margin-top: 0;">📊 Overall Assessment</h3>
      <p style="color: var(--text-neutral); line-height: 1.6; margin: 0;">${result.summary}</p>
    </div>
    ${missingItemsHTML}
    ${strengthsHTML}
    ${tipsHTML}
  `, [
    { label: 'Close', onclick: 'closeModal()' }
  ]);
  showToast('Wardrobe analysis complete!', 'success');
}

window.createOutfitFromSuggestion = function(itemIds) {
  // Save selected items to session storage and redirect
  sessionStorage.setItem('suggested_outfit_items', JSON.stringify(itemIds));
  window.location.href = 'outfit-builder.html?suggested=true';
};

// ============================================================================
// RENDER SHARED OUTFIT PAGE
// ============================================================================

function renderSharedOutfit(shared) {
  const outfit = shared.outfit;
  const items = outfit.itemIds.map(id => appData.items.find(i => i.id == id)).filter(Boolean);

  // Update page title
  document.title = `${outfit.name} - Shared Outfit - go-threads`;

  // Render outfit preview (simplified for share page)
  const canvas = document.querySelector('.share-canvas');
  if (canvas && outfit.positions) {
    canvas.innerHTML = outfit.positions.map(pos => {
      const item = items.find(i => i.id === pos.id);
      if (!item) return '';
      return `
        <div class="canvas-item" style="left: ${pos.x}px; top: ${pos.y}px; z-index: ${pos.z};">
          <img src="${item.image}" alt="${item.name}">
        </div>
      `;
    }).join('');
  }

  // Update outfit info
  const titleEl = document.querySelector('.share-title');
  if (titleEl) titleEl.textContent = outfit.name;

  const metaEl = document.querySelector('.share-meta');
  if (metaEl) metaEl.textContent = `${items.length} items • Worn ${outfit.wearCount || 0} times`;

  // Update rating
  const avgRating = shared.ratings && shared.ratings.length > 0
    ? shared.ratings.reduce((sum, r) => sum + r.rating, 0) / shared.ratings.length
    : 0;

  const ratingDisplay = document.querySelector('.rating-display');
  if (ratingDisplay) {
    const stars = '★'.repeat(Math.round(avgRating)) + '☆'.repeat(5 - Math.round(avgRating));
    ratingDisplay.innerHTML = `
      <span class="rating-stars">${stars}</span>
      <span class="rating-count">${avgRating.toFixed(1)} (${shared.ratings?.length || 0} ratings)</span>
    `;
  }

  // Update items list
  const itemsContainer = document.querySelector('.share-items');
  if (itemsContainer) {
    itemsContainer.innerHTML = items.map(item => `
      <div class="share-item">
        <img src="${item.image}" alt="${item.name}">
        <div class="share-item-name">${item.name}<br>$${item.price.toFixed(2)}</div>
      </div>
    `).join('');
  }

  // Setup rating button
  const rateBtn = document.querySelector('.share-actions button');
  if (rateBtn) {
    rateBtn.onclick = () => {
      const rating = prompt('Rate this outfit (1-5 stars):');
      if (rating && rating >= 1 && rating <= 5) {
        rateSharedOutfit(shared.shareId, parseInt(rating));
      }
    };
  }
}

// Export additional functions for HTML onclick handlers
window.changeMonth = changeMonth;
window.openDayModal = openDayModal;
window.saveCalendarOutfit = saveCalendarOutfit;
window.getAICalendarSuggestionsForDay = getAICalendarSuggestionsForDay;
window.shareOutfit = shareOutfit;
window.rateSharedOutfit = rateSharedOutfit;
window.deleteOutfit = deleteOutfit;
window.confirmDeleteOutfit = confirmDeleteOutfit;
window.duplicateOutfit = duplicateOutfit;
window.viewOutfitDetail = viewOutfitDetail;
window.editOutfit = editOutfit;
window.renderSharedOutfit = renderSharedOutfit;

// ============================================================================
// RETRY AI ANALYSIS FOR ITEMS
// ============================================================================

async function retryAIFromWardrobe(itemId) {
  const container = document.getElementById(`retry-container-${itemId}`);
  if (container) {
    container.innerHTML = `
      <div style="width: 100%; margin-top: 0.5rem; padding: 0.5rem; text-align: center;">
        <div style="font-size: 0.75rem; color: var(--text-neutral); margin-bottom: 0.25rem;">⏳ Analyzing...</div>
        <div style="width: 100%; height: 4px; background: var(--border); border-radius: 2px; overflow: hidden;">
          <div id="retry-progress-${itemId}" style="width: 0%; height: 100%; background: var(--primary); transition: width 0.3s;"></div>
        </div>
      </div>
    `;
    // Animate progress
    let progress = 0;
    const bar = document.getElementById(`retry-progress-${itemId}`);
    const interval = setInterval(() => {
      progress += 2;
      if (bar) bar.style.width = `${Math.min(progress, 90)}%`;
      if (progress >= 90) clearInterval(interval);
    }, 100);
  }

  try {
    await retryAIAnalysis(itemId);
  } catch (e) {
    // retryAIAnalysis handles its own errors
  }
}
window.retryAIFromWardrobe = retryAIFromWardrobe;

async function retryAIAnalysis(itemId) {
  if (!aiAvailable) {
    showToast('AI not available. Start the API server.', 'error');
    return;
  }

  try {
    // Fetch item from API
    const response = await fetch(`http://localhost:8556/api/items/${itemId}`);
    if (!response.ok) {
      throw new Error('Item not found');
    }
    const item = await response.json();

    // Convert image URL to base64 if needed
    let imageData = item.image_url;
    if (!imageData.startsWith('data:')) {
      const imgResponse = await fetch(imageData);
      const blob = await imgResponse.blob();
      imageData = await new Promise((resolve) => {
        const reader = new FileReader();
        reader.onloadend = () => resolve(reader.result);
        reader.readAsDataURL(blob);
      });
    }

    const analysis = await analyzeImageWithAI(imageData);

    if (analysis) {
      // Prepare updated item data
      const updatedItem = {
        name: item.name,
        description: analysis.description || item.description || '',
        category: analysis.category || item.category,
        price: item.price,
        color: analysis.color || item.color,
        brand: analysis.brand || item.brand,
        ai_analysis: analysis.raw_response || analysis.description || 'AI analysis completed',
        tags: (analysis.tags && analysis.tags.length > 0) ? analysis.tags : item.tags
      };

      // Update name if it's still default
      if ((item.name === 'Unnamed Item' || !item.name) && (analysis.name || analysis.description)) {
        updatedItem.name = analysis.name || analysis.description.substring(0, 50);
      }

      // Update category if it's still default
      if ((item.category === 'Uncategorized' || !item.category) && analysis.category && analysis.category !== 'Unknown') {
        updatedItem.category = analysis.category;
      }

      // Update via API
      await fetch(`http://localhost:8556/api/items/${itemId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updatedItem)
      });

      showToast('AI analysis completed!', 'success');

      // Reload the page to show updated info
      setTimeout(() => {
        window.location.reload();
      }, 1000);
    } else {
      showToast('AI analysis failed. Please try again.', 'error');
    }
  } catch (error) {
    console.error('Retry analysis error:', error);
    showToast('Error during AI analysis: ' + error.message, 'error');
  }
}

// ============================================================================
// VOICE INPUT SUPPORT
// ============================================================================

let voiceRecognition = null;
let isRecording = false;

function initVoiceInput() {
  // Check for browser support
  const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
  
  if (!SpeechRecognition) {
    console.log('Speech recognition not supported in this browser');
    return false;
  }

  voiceRecognition = new SpeechRecognition();
  voiceRecognition.continuous = false;
  voiceRecognition.interimResults = false;
  voiceRecognition.lang = 'en-US';

  voiceRecognition.onresult = (event) => {
    const transcript = event.results[0][0].transcript;
    handleVoiceInput(transcript);
  };

  voiceRecognition.onerror = (event) => {
    console.error('Speech recognition error:', event.error);
    isRecording = false;
    updateVoiceButtonState(false);
    
    if (event.error === 'not-allowed') {
      showToast('Microphone access denied. Please enable in browser settings.', 'error');
    } else {
      showToast('Voice input failed: ' + event.error, 'error');
    }
  };

  voiceRecognition.onend = () => {
    isRecording = false;
    updateVoiceButtonState(false);
  };

  return true;
}

function startVoiceInput() {
  if (!voiceRecognition) {
    const supported = initVoiceInput();
    if (!supported) {
      showToast('Voice input not supported in this browser. Try Chrome or Edge.', 'error');
      return;
    }
  }

  if (isRecording) {
    stopVoiceInput();
    return;
  }

  try {
    voiceRecognition.start();
    isRecording = true;
    updateVoiceButtonState(true);
    showToast('🎤 Listening... Speak now', 'info');
  } catch (error) {
    console.error('Failed to start voice input:', error);
    showToast('Failed to start microphone', 'error');
  }
}

function stopVoiceInput() {
  if (voiceRecognition && isRecording) {
    voiceRecognition.stop();
    isRecording = false;
    updateVoiceButtonState(false);
  }
}

function updateVoiceButtonState(recording) {
  const voiceBtns = document.querySelectorAll('.voice-input-btn');
  voiceBtns.forEach(btn => {
    if (recording) {
      btn.classList.add('recording');
      btn.innerHTML = '🔴';
      btn.title = 'Recording... Click to stop';
    } else {
      btn.classList.remove('recording');
      btn.innerHTML = '🎤';
      btn.title = 'Voice input';
    }
  });
}

function handleVoiceInput(transcript) {
  console.log('Voice transcript:', transcript);
  showToast(`Heard: "${transcript}"`, 'success');

  // Check which page we're on and handle appropriately
  const chatInput = document.getElementById('chat-input');
  if (chatInput) {
    // In chat interface
    chatInput.value = transcript;
    const sendBtn = document.getElementById('chat-send-btn');
    if (sendBtn) sendBtn.click();
    return;
  }

  // Parse voice commands for other pages
  processVoiceCommand(transcript);
}

async function processVoiceCommand(transcript) {
  const lower = transcript.toLowerCase();

  // Upload/Add item commands
  if (lower.includes('add') || lower.includes('upload') || lower.includes('new item')) {
    window.location.href = 'upload.html';
    return;
  }

  // Create outfit commands
  if (lower.includes('create outfit') || lower.includes('make outfit')) {
    window.location.href = 'outfit-builder.html';
    return;
  }

  // Search/Filter commands
  if (lower.includes('show') || lower.includes('find') || lower.includes('search')) {
    // Extract category or color
    const categories = ['tops', 'bottoms', 'shoes', 'outerwear', 'accessories'];
    const colors = ['blue', 'red', 'white', 'black', 'green', 'navy', 'beige'];
    
    for (const cat of categories) {
      if (lower.includes(cat)) {
        const capitalizedCat = cat.charAt(0).toUpperCase() + cat.slice(0, -1);
        window.location.href = `wardrobe.html?category=${capitalizedCat}`;
        return;
      }
    }

    for (const color of colors) {
      if (lower.includes(color)) {
        window.location.href = `wardrobe.html?search=${color}`;
        return;
      }
    }

    // Default to wardrobe
    window.location.href = 'wardrobe.html';
    return;
  }

  // Analytics commands
  if (lower.includes('stats') || lower.includes('analytics') || lower.includes('cost per wear')) {
    window.location.href = 'analytics.html';
    return;
  }

  // Calendar commands
  if (lower.includes('calendar') || lower.includes('what should i wear')) {
    window.location.href = 'calendar.html';
    return;
  }

  // AI Stylist commands
  if (lower.includes('suggest') || lower.includes('recommend') || lower.includes('what goes with')) {
    window.location.href = 'outfits.html';
    return;
  }

  // If we couldn't parse the command, show help
  showToast('Try: "Show my tops", "Create an outfit", "Add new item"', 'info');
}

