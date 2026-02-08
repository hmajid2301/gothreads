// Throwaway mockup JavaScript - to be replaced with HTMX + Alpine

// Sample data
const mockItems = [
  { id: 1, name: 'Blue Denim Jeans', category: 'Bottoms', price: 79.99, image: 'https://via.placeholder.com/300/4A90E2/ffffff?text=Blue+Jeans' },
  { id: 2, name: 'White T-Shirt', category: 'Tops', price: 29.99, image: 'https://via.placeholder.com/300/F5F5F5/333333?text=White+Tee' },
  { id: 3, name: 'Black Leather Jacket', category: 'Outerwear', price: 199.99, image: 'https://via.placeholder.com/300/000000/ffffff?text=Leather+Jacket' },
  { id: 4, name: 'Red Sneakers', category: 'Shoes', price: 89.99, image: 'https://via.placeholder.com/300/E74C3C/ffffff?text=Sneakers' },
  { id: 5, name: 'Gray Hoodie', category: 'Tops', price: 59.99, image: 'https://via.placeholder.com/300/95A5A6/ffffff?text=Hoodie' },
  { id: 6, name: 'Navy Chinos', category: 'Bottoms', price: 69.99, image: 'https://via.placeholder.com/300/34495E/ffffff?text=Chinos' },
];

// Store uploaded items and outfits
let uploadedItems = [...mockItems];
let currentOutfit = [];

// Upload handling
function handleFileUpload(event) {
  const files = event.target.files;
  if (!files || files.length === 0) return;

  const file = files[0];
  const reader = new FileReader();

  reader.onload = function(e) {
    const newItem = {
      id: Date.now(),
      name: file.name.replace(/\.[^/.]+$/, ''),
      category: 'Uncategorized',
      price: 0,
      image: e.target.result
    };

    uploadedItems.push(newItem);
    showToast('Item uploaded! Add details below.');

    // Redirect to item detail page
    setTimeout(() => {
      window.location.href = `item-detail.html?id=${newItem.id}`;
    }, 1000);
  };

  reader.readAsDataURL(file);
}

// Drag and drop for outfit canvas
let draggedElement = null;
let offsetX = 0;
let offsetY = 0;

function makeDraggable(element) {
  element.addEventListener('mousedown', startDrag);
}

function startDrag(e) {
  draggedElement = e.currentTarget;
  const rect = draggedElement.getBoundingClientRect();
  const canvas = document.querySelector('.outfit-canvas');
  const canvasRect = canvas.getBoundingClientRect();

  offsetX = e.clientX - rect.left;
  offsetY = e.clientY - rect.top;

  draggedElement.classList.add('selected');

  document.addEventListener('mousemove', drag);
  document.addEventListener('mouseup', stopDrag);
}

function drag(e) {
  if (!draggedElement) return;

  const canvas = document.querySelector('.outfit-canvas');
  const canvasRect = canvas.getBoundingClientRect();

  let x = e.clientX - canvasRect.left - offsetX;
  let y = e.clientY - canvasRect.top - offsetY;

  // Constrain to canvas bounds
  x = Math.max(0, Math.min(x, canvasRect.width - draggedElement.offsetWidth));
  y = Math.max(0, Math.min(y, canvasRect.height - draggedElement.offsetHeight));

  draggedElement.style.left = x + 'px';
  draggedElement.style.top = y + 'px';
}

function stopDrag() {
  if (draggedElement) {
    draggedElement.classList.remove('selected');
    draggedElement = null;
  }
  document.removeEventListener('mousemove', drag);
  document.removeEventListener('mouseup', stopDrag);
}

// Add item to outfit canvas
function addToCanvas(item) {
  const canvas = document.querySelector('.outfit-canvas');
  if (!canvas) return;

  const canvasItem = document.createElement('div');
  canvasItem.className = 'canvas-item';
  canvasItem.dataset.itemId = item.id;
  canvasItem.style.left = (Math.random() * 200 + 100) + 'px';
  canvasItem.style.top = (Math.random() * 200 + 100) + 'px';

  const img = document.createElement('img');
  img.src = item.image;
  img.alt = item.name;

  canvasItem.appendChild(img);
  canvas.appendChild(canvasItem);

  makeDraggable(canvasItem);

  if (!currentOutfit.find(i => i.id === item.id)) {
    currentOutfit.push(item);
  }

  showToast(`Added ${item.name} to outfit`);
}

// Save outfit
function saveOutfit() {
  if (currentOutfit.length === 0) {
    showToast('Add items to your outfit first!');
    return;
  }

  // In real app, this would POST to backend
  console.log('Saving outfit:', currentOutfit);

  const outfitName = prompt('Name your outfit:');
  if (outfitName) {
    showToast(`Outfit "${outfitName}" saved!`);
    setTimeout(() => {
      window.location.href = 'wardrobe.html';
    }, 1500);
  }
}

// Toast notifications
function showToast(message) {
  const toast = document.getElementById('toast');
  if (!toast) {
    const newToast = document.createElement('div');
    newToast.id = 'toast';
    newToast.className = 'toast';
    document.body.appendChild(newToast);
  }

  const toastElement = document.getElementById('toast');
  toastElement.textContent = message;
  toastElement.classList.add('show');

  setTimeout(() => {
    toastElement.classList.remove('show');
  }, 3000);
}

// Render item grid
function renderItemGrid(items, containerId) {
  const container = document.getElementById(containerId);
  if (!container) return;

  container.innerHTML = '';

  items.forEach(item => {
    const card = document.createElement('div');
    card.className = 'item-card';
    card.onclick = () => window.location.href = `item-detail.html?id=${item.id}`;

    card.innerHTML = `
      <img src="${item.image}" alt="${item.name}" class="item-image">
      <div class="item-info">
        <div class="item-name">${item.name}</div>
        <div class="item-meta">${item.category} • $${item.price.toFixed(2)}</div>
      </div>
    `;

    container.appendChild(card);
  });
}

// Render item picker (for outfit builder)
function renderItemPicker(items, containerId) {
  const container = document.getElementById(containerId);
  if (!container) return;

  container.innerHTML = '<h3>Your Items</h3>';

  items.forEach(item => {
    const pickerItem = document.createElement('div');
    pickerItem.className = 'picker-item';
    pickerItem.onclick = () => addToCanvas(item);

    pickerItem.innerHTML = `
      <img src="${item.image}" alt="${item.name}">
      <div class="picker-item-info">
        <div class="picker-item-name">${item.name}</div>
        <div class="picker-item-category">${item.category}</div>
      </div>
    `;

    container.appendChild(pickerItem);
  });
}

// Get item by ID from URL
function getItemFromURL() {
  const params = new URLSearchParams(window.location.search);
  const id = parseInt(params.get('id'));
  return uploadedItems.find(item => item.id === id);
}

// Initialize based on page
document.addEventListener('DOMContentLoaded', function() {
  // Upload page
  const uploadInput = document.getElementById('upload-input');
  if (uploadInput) {
    uploadInput.addEventListener('change', handleFileUpload);
  }

  // Wardrobe grid
  renderItemGrid(uploadedItems, 'wardrobe-grid');

  // Outfit builder
  renderItemPicker(uploadedItems, 'item-picker');

  // Save outfit button
  const saveBtn = document.getElementById('save-outfit');
  if (saveBtn) {
    saveBtn.addEventListener('click', saveOutfit);
  }

  // Item detail page
  const itemDetail = document.getElementById('item-detail');
  if (itemDetail) {
    const item = getItemFromURL();
    if (item) {
      document.getElementById('item-name').value = item.name;
      document.getElementById('item-category').value = item.category;
      document.getElementById('item-price').value = item.price;
      document.getElementById('item-image-preview').src = item.image;
    }
  }
});
