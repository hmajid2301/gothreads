-- +goose Up
-- Initial schema for gothreads

-- Users table (for future OAuth integration)
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Clothing items table
CREATE TABLE IF NOT EXISTS items (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(50) NOT NULL, -- Tops, Bottoms, Shoes, Outerwear, Accessories
    price DECIMAL(10, 2) DEFAULT 0,
    brand VARCHAR(100),
    color VARCHAR(50),
    season VARCHAR(20), -- All Season, Spring, Summer, Fall, Winter
    image_url TEXT NOT NULL,
    notes TEXT,
    wear_count INTEGER DEFAULT 0,
    ai_analysis TEXT, -- JSON blob with AI analysis results
    tags TEXT[], -- Array of tags
    uploaded_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Outfits table
CREATE TABLE IF NOT EXISTS outfits (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    notes TEXT,
    wear_count INTEGER DEFAULT 0,
    rating DECIMAL(3, 2), -- 0.00 to 5.00
    body_image_url TEXT, -- Optional custom body/mannequin image
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Outfit items (junction table for many-to-many)
CREATE TABLE IF NOT EXISTS outfit_items (
    id BIGSERIAL PRIMARY KEY,
    outfit_id BIGINT REFERENCES outfits(id) ON DELETE CASCADE,
    item_id BIGINT REFERENCES items(id) ON DELETE CASCADE,
    position_x INTEGER DEFAULT 0,
    position_y INTEGER DEFAULT 0,
    position_z INTEGER DEFAULT 0, -- Layering order
    scale DECIMAL(3, 2) DEFAULT 1.0,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(outfit_id, item_id)
);

-- Wear history table
CREATE TABLE IF NOT EXISTS wear_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    item_id BIGINT REFERENCES items(id) ON DELETE CASCADE,
    outfit_id BIGINT REFERENCES outfits(id) ON DELETE SET NULL,
    worn_at TIMESTAMP DEFAULT NOW()
);

-- Calendar events table
CREATE TABLE IF NOT EXISTS calendar_events (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    event_date DATE NOT NULL,
    event_name VARCHAR(255),
    weather VARCHAR(50),
    outfit_id BIGINT REFERENCES outfits(id) ON DELETE SET NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Shared outfits table
CREATE TABLE IF NOT EXISTS shared_outfits (
    id BIGSERIAL PRIMARY KEY,
    outfit_id BIGINT REFERENCES outfits(id) ON DELETE CASCADE,
    share_id VARCHAR(8) UNIQUE NOT NULL, -- Short random ID for URL
    created_at TIMESTAMP DEFAULT NOW(),
    views INTEGER DEFAULT 0
);

-- Outfit ratings (for shared outfits)
CREATE TABLE IF NOT EXISTS outfit_ratings (
    id BIGSERIAL PRIMARY KEY,
    outfit_id BIGINT REFERENCES outfits(id) ON DELETE CASCADE,
    rating INTEGER CHECK (rating >= 1 AND rating <= 5),
    rated_at TIMESTAMP DEFAULT NOW()
);

-- User settings table
CREATE TABLE IF NOT EXISTS user_settings (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    auto_remove_bg BOOLEAN DEFAULT TRUE,
    auto_tag BOOLEAN DEFAULT TRUE,
    ai_provider VARCHAR(50) DEFAULT 'local',
    currency VARCHAR(3) DEFAULT 'USD',
    date_format VARCHAR(20) DEFAULT 'MM/DD/YYYY',
    max_wears_before_laundry INTEGER DEFAULT 3,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_items_user_id ON items(user_id);
CREATE INDEX IF NOT EXISTS idx_items_category ON items(category);
CREATE INDEX IF NOT EXISTS idx_outfits_user_id ON outfits(user_id);
CREATE INDEX IF NOT EXISTS idx_outfit_items_outfit_id ON outfit_items(outfit_id);
CREATE INDEX IF NOT EXISTS idx_outfit_items_item_id ON outfit_items(item_id);
CREATE INDEX IF NOT EXISTS idx_wear_history_user_id ON wear_history(user_id);
CREATE INDEX IF NOT EXISTS idx_wear_history_item_id ON wear_history(item_id);
CREATE INDEX IF NOT EXISTS idx_calendar_events_user_id ON calendar_events(user_id);
CREATE INDEX IF NOT EXISTS idx_calendar_events_date ON calendar_events(event_date);
CREATE INDEX IF NOT EXISTS idx_shared_outfits_share_id ON shared_outfits(share_id);

-- +goose Down
DROP TABLE IF EXISTS outfit_ratings CASCADE;
DROP TABLE IF EXISTS shared_outfits CASCADE;
DROP TABLE IF EXISTS calendar_events CASCADE;
DROP TABLE IF EXISTS wear_history CASCADE;
DROP TABLE IF NOT EXISTS outfit_items CASCADE;
DROP TABLE IF EXISTS outfits CASCADE;
DROP TABLE IF EXISTS items CASCADE;
DROP TABLE IF EXISTS user_settings CASCADE;
DROP TABLE IF EXISTS users CASCADE;
