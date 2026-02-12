-- name: GetItem :one
SELECT * FROM items
WHERE id = $1 AND user_id = $2 LIMIT 1;

-- name: ListItems :many
SELECT * FROM items
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: ListItemsByCategory :many
SELECT * FROM items
WHERE user_id = $1 AND category = $2
ORDER BY created_at DESC;

-- name: CreateItem :one
INSERT INTO items (
  user_id, name, category, price, brand, color, season,
  image_url, notes, wear_count, ai_analysis, tags
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
)
RETURNING *;

-- name: UpdateItem :one
UPDATE items
SET name = $2,
    category = $3,
    price = $4,
    brand = $5,
    color = $6,
    season = $7,
    notes = $8,
    ai_analysis = $9,
    tags = $10,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteItem :exec
DELETE FROM items
WHERE id = $1 AND user_id = $2;

-- name: IncrementWearCount :exec
UPDATE items
SET wear_count = wear_count + 1
WHERE id = $1;
