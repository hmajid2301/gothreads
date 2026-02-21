// go-threads Service Worker
// Implements offline-first caching strategy for PWA functionality

const CACHE_NAME = 'go-threads-v1';
const CACHE_VERSION = 1;

// Essential app resources for offline functionality
const STATIC_CACHE_URLS = [
  '/',
  '/index.html',
  '/wardrobe.html', 
  '/outfits.html',
  '/outfit-builder.html',
  '/settings.html',
  '/css/style.css',
  '/js/app.js',
  '/manifest.json'
];

// Dynamic cache for wardrobe images and API data
const IMAGE_CACHE_NAME = 'go-threads-images-v1';
const DATA_CACHE_NAME = 'go-threads-data-v1';

// Install event - cache essential resources
self.addEventListener('install', (event) => {
  console.log('[SW] Installing service worker...');
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => {
        console.log('[SW] Caching app shell');
        return cache.addAll(STATIC_CACHE_URLS);
      })
      .then(() => self.skipWaiting())
  );
});

// Activate event - cleanup old caches
self.addEventListener('activate', (event) => {
  console.log('[SW] Activating service worker...');
  event.waitUntil(
    caches.keys()
      .then((cacheNames) => {
        return Promise.all(
          cacheNames.map((cacheName) => {
            if (cacheName !== CACHE_NAME && 
                cacheName !== IMAGE_CACHE_NAME && 
                cacheName !== DATA_CACHE_NAME) {
              console.log('[SW] Deleting old cache:', cacheName);
              return caches.delete(cacheName);
            }
          })
        );
      })
      .then(() => self.clients.claim())
  );
});

// Fetch event - implement offline-first strategy
self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);
  
  // Handle API requests with network-first strategy
  if (url.pathname.startsWith('/api/')) {
    event.respondWith(networkFirstStrategy(request));
    return;
  }
  
  // Handle image requests with cache-first strategy
  if (request.destination === 'image' || url.pathname.match(/\.(jpg|jpeg|png|gif|webp|svg)$/i)) {
    event.respondWith(cacheFirstStrategy(request, IMAGE_CACHE_NAME));
    return;
  }
  
  // Handle app shell and static resources with cache-first strategy
  if (STATIC_CACHE_URLS.includes(url.pathname) || url.pathname === '/') {
    event.respondWith(cacheFirstStrategy(request, CACHE_NAME));
    return;
  }
  
  // Default to network-first for everything else
  event.respondWith(networkFirstStrategy(request));
});

// Cache-first strategy: Check cache first, fallback to network
async function cacheFirstStrategy(request, cacheName = CACHE_NAME) {
  try {
    const cache = await caches.open(cacheName);
    const cachedResponse = await cache.match(request);
    
    if (cachedResponse) {
      console.log('[SW] Cache hit:', request.url);
      return cachedResponse;
    }
    
    console.log('[SW] Cache miss, fetching:', request.url);
    const networkResponse = await fetch(request);
    
    // Cache successful responses
    if (networkResponse.ok) {
      const responseClone = networkResponse.clone();
      cache.put(request, responseClone);
    }
    
    return networkResponse;
  } catch (error) {
    console.error('[SW] Cache-first strategy failed:', error);
    throw error;
  }
}

// Network-first strategy: Try network first, fallback to cache
async function networkFirstStrategy(request) {
  try {
    const networkResponse = await fetch(request);
    
    // Cache successful API responses for offline access
    if (networkResponse.ok && request.url.includes('/api/')) {
      const cache = await caches.open(DATA_CACHE_NAME);
      const responseClone = networkResponse.clone();
      cache.put(request, responseClone);
    }
    
    return networkResponse;
  } catch (error) {
    console.log('[SW] Network failed, checking cache:', request.url);
    
    // Try to serve from cache if network fails
    const cache = await caches.open(DATA_CACHE_NAME);
    const cachedResponse = await cache.match(request);
    
    if (cachedResponse) {
      console.log('[SW] Serving from cache offline:', request.url);
      return cachedResponse;
    }
    
    // Return offline page for navigation requests
    if (request.mode === 'navigate') {
      const offlineCache = await caches.open(CACHE_NAME);
      return offlineCache.match('/index.html');
    }
    
    throw error;
  }
}

// Background sync event - sync wardrobe data when online
self.addEventListener('sync', (event) => {
  console.log('[SW] Background sync triggered:', event.tag);
  
  if (event.tag === 'wardrobe-sync') {
    event.waitUntil(syncWardrobeData());
  }
});

// Sync wardrobe data function
async function syncWardrobeData() {
  try {
    console.log('[SW] Syncing wardrobe data...');
    
    // Check if we have any pending uploads in IndexedDB
    const pendingUploads = await getPendingUploads();
    
    for (const upload of pendingUploads) {
      try {
        const response = await fetch('/api/items', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(upload.data)
        });
        
        if (response.ok) {
          await removePendingUpload(upload.id);
          console.log('[SW] Synced upload:', upload.id);
        }
      } catch (error) {
        console.error('[SW] Failed to sync upload:', upload.id, error);
      }
    }
    
    // Refresh wardrobe data cache
    const cache = await caches.open(DATA_CACHE_NAME);
    cache.delete('/api/items');
    
    console.log('[SW] Wardrobe sync complete');
  } catch (error) {
    console.error('[SW] Background sync failed:', error);
  }
}

// Helper functions for IndexedDB operations (simplified for prototype)
async function getPendingUploads() {
  // In a real implementation, this would read from IndexedDB
  // For now, return empty array as we're focusing on cache strategy
  return [];
}

async function removePendingUpload(id) {
  // In a real implementation, this would remove from IndexedDB
  console.log('[SW] Would remove pending upload:', id);
}

// Message handling for client communication
self.addEventListener('message', (event) => {
  console.log('[SW] Message received:', event.data);
  
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }
  
  if (event.data && event.data.type === 'GET_VERSION') {
    event.ports[0].postMessage({ version: CACHE_VERSION });
  }
});